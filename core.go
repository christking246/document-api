package main

import (
	"documentApi/data"
	"documentApi/utils"
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// TODO: some of these regex (\w)s should probably be [a-zA-Z0-9_] or similar
// TODO: It's important to note that code that is commented out in the target files will still be parsed as active...this should probably be fixed
// TODO: consider parsing the top level route path for a given controller if exists

const DefaultAuth = "DocsToken"

var functionRegex = regexp.MustCompile(`\[Function\((?:nameof\()?"?(?<fname>\w+)"?\)?\)\]`)
var authenticationRegex = regexp.MustCompile(`\[Require(?<type>DocsTokenGroups|S2SToken|DocsToken|PlatformApiAuth)(?:\((?<groups>[^)]*)\))?\]`)

// TODO: parse the whole function/method body?
var classFunctionRegex = regexp.MustCompile(`(?:public|protected|private)\s+(?:async\s+)?[\w<>]+\s+(?<fname>\w+)\(`)
var httpTriggerRegex = regexp.MustCompile(`\[HttpTrigger\([\w.]+,\s*(?<methods>"\w+"\s*,\s*)+\s*Route\s*=\s*(?:"(?<route>[^"]+)"|(?<route>[^])]+))\)\]`)
var timeTrigger = regexp.MustCompile(`\[TimerTrigger\("(?<cron>[^"]+)"[^)]*\)\]`)

var pathVarRegex = regexp.MustCompile(`{([a-zA-Z0-9_]+)}`)

// read host json file to get the path prepended to all endpoints in a given package
func getApiPrefixes(repoPath string, logger *logrus.Logger) map[string]string {
	jsonEntries, err := utils.GetFiles(repoPath, []string{".json"}, false, true, true)
	if err != nil {
		logger.Warn("Error reading repo '" + repoPath + "' to locate host json file: " + err.Error())
	}

	// TODO: write array filtering function?
	prefixes := make(map[string]string)
	for _, entry := range jsonEntries {
		if entry.Name == "host.json" {
			var prefixKey = utils.Base(utils.Dir(entry.Path))
			hostFileData, err := os.ReadFile(entry.Path)
			if err != nil {
				logger.Warn("Error reading host json file: " + entry.Path + ": " + err.Error())
				continue
			}
			var hostData data.ApiMetaData
			if err := json.Unmarshal(hostFileData, &hostData); err != nil {
				logger.Warn("Error parsing host file: " + entry.Path + ": " + err.Error())
				continue
			}
			prefixes[prefixKey] = hostData.Extensions.Http.RoutePrefix
		}
	}

	return prefixes
}

func parseFunctionHeader(str string, endpoint *data.EndpointMetaData) int {
	var m = classFunctionRegex.FindStringSubmatchIndex(str)[0]
	var begin = m + len(classFunctionRegex.FindStringSubmatch(str)[0])
	var stop = len(str)
	var openers = 0
	var skipped = 0

	for i := begin; i < len(str); i++ {
		if str[i] == '(' {
			openers++
		}
		if str[i] == ')' {
			if openers == 0 {
				stop = i
				break
			}
			openers--
		}
		if str[i] == '\n' {
			skipped++
		}
	}

	str = str[begin:stop]

	// get the trigger type
	// Maybe this should just be set from the individual trigger if statements (but there are only 2 atm)
	// TODO: move this regex so it is only created once
	var triggerKeys = make([]string, 0, len(data.TriggerType))
	for k := range data.TriggerType {
		triggerKeys = append(triggerKeys, k)
	}
	var triggerTypeRegex = regexp.MustCompile(`\[(?<triggerType>` + strings.Join(triggerKeys, "|") + `)Trigger[\S\s]*?\]`)
	triggerTypeMatch := triggerTypeRegex.FindStringSubmatch(str)
	if len(triggerTypeMatch) > 0 {
		endpoint.TriggerType = data.TriggerType[triggerTypeMatch[1]]
	} else {
		endpoint.TriggerType = data.TriggerType["UNKNOWN"]
	}

	// pull out the route and path vars (if any)
	httpTriggerRegexMatch := httpTriggerRegex.FindStringSubmatch(str)
	if len(httpTriggerRegexMatch) > 0 {
		// TODO: write function to split by regex
		var noSpaces = strings.ReplaceAll(httpTriggerRegexMatch[1][0:len(httpTriggerRegexMatch[1])-2], " ", "")
		var noQuotes = strings.ReplaceAll(noSpaces, "\"", "")
		endpoint.Methods = strings.Split(noQuotes, ",")
		endpoint.Route = httpTriggerRegexMatch[2] // this will not resolve paths that are built from variables
		if len(endpoint.Route) < 1 && len(httpTriggerRegexMatch) > 3 {
			// TODO: consider replacing "Route=null" with empty string or making as an inaccessible path
			endpoint.Route = httpTriggerRegexMatch[3] // blame go for not having named capture groups
		}
		var matches = pathVarRegex.FindAllStringSubmatch(endpoint.Route, -1)
		if len(matches) > 0 {
			for _, match := range matches {
				endpoint.PathParameters = append(endpoint.PathParameters, match[1])
			}
		}
	}

	// pull out the cron expression if this is a time trigger
	timeTriggerMatch := timeTrigger.FindStringSubmatch(str)
	if len(timeTriggerMatch) > 0 {
		endpoint.Interval = timeTriggerMatch[1]
	}

	if skipped > 0 {
		return skipped - 1
	}
	return skipped
}

// TODO: break this up into smaller functions to write separate unit tests for each?
func parse(targetFile data.FileMetaData, logger *logrus.Logger) []data.EndpointMetaData {
	fileData, err := os.ReadFile(targetFile.Path)
	if err != nil {
		logger.Error("Error reading file: " + targetFile.Path + ": " + err.Error())
		return []data.EndpointMetaData{}
	}

	var fileDataString string = string(fileData)
	var functionMatches = functionRegex.FindAllStringSubmatch(fileDataString, -1)
	if len(functionMatches) == 0 {
		logger.Debug("No functions found in file: " + targetFile.Path)
		return []data.EndpointMetaData{}
	}
	logger.Debug("Found " + strconv.Itoa(len(functionMatches)) + " functions in file: " + targetFile.Path)

	fileDataString = strings.ReplaceAll(fileDataString, "\r\n", "\n") // In case windows is using \r\n
	lines := strings.Split(fileDataString, "\n")                      // TODO: this works in linux, but will it in windows?

	var endpoints = []data.EndpointMetaData{}
	var currentEndpoint data.EndpointMetaData
	var first = true
	var runningLength = 0

	for lineNumber := 0; lineNumber < len(lines); lineNumber++ { // need this be closer to a counter for loop, because I will probably need to jump through lines
		line := lines[lineNumber]

		// when you encounter the function header, no more decorators to parse for this function
		// save the collected metadata so far
		var classFunctionMatch = classFunctionRegex.FindStringSubmatch(line)
		if len(classFunctionMatch) > 0 {
			if first {
				first = false
			} else {
				// if this function has no name, it's probably a regular function/method
				if len(currentEndpoint.Name) > 1 {
					var skipped = parseFunctionHeader(fileDataString[runningLength:], &currentEndpoint)
					for i := lineNumber; i < lineNumber+skipped-1; i++ {
						runningLength += len(lines[i])
					}
					lineNumber += skipped
					endpoints = append(endpoints, currentEndpoint)
				}
				currentEndpoint = data.EndpointMetaData{}
			}
		}

		// function name
		var functionMatch = functionRegex.FindStringSubmatch(line)
		if len(functionMatch) > 0 {
			currentEndpoint.Name = functionMatch[1]
		}

		// authentication
		var authenticationMatch = authenticationRegex.FindStringSubmatch(line)
		if len(authenticationMatch) > 1 {
			if len(authenticationMatch) > 2 && len(authenticationMatch[2]) > 0 {
				// TODO: write function to split by regex
				var noCommaSpace = strings.ReplaceAll(authenticationMatch[2], ", ", ",")
				var noSpace = strings.ReplaceAll(noCommaSpace, " ", ",")
				var noQuotes = strings.ReplaceAll(noSpace, "\"", "") // this will make it hard to determine if is a docs token group vs "arbitrary string" ... if that's a concern
				var modes = strings.Split(noQuotes, ",")
				currentEndpoint.Authentication = append(currentEndpoint.Authentication, modes...)
			} else {
				currentEndpoint.Authentication = append(currentEndpoint.Authentication, authenticationMatch[1])
			}
		}

		runningLength += len(line) // this can probably added in the 'for' header
	}

	if len(currentEndpoint.Name) > 0 {
		endpoints = append(endpoints, currentEndpoint)
	}

	if len(functionMatches) != len(endpoints) {
		// should this be a hard fail?
		logger.Warn("Error parsing file '" + targetFile.Path + "'. Documented " + strconv.Itoa(len(endpoints)) + " functions, but expected " + strconv.Itoa(len(functionMatches)))
	}

	return endpoints
}
