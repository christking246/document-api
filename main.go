package main

import (
	"documentApi/data"
	"documentApi/documenters"
	"documentApi/utils"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// TODO: add option to keep old vars (env, path params, etc) from existing collections upon updating
// TODO: add option to create documentation for a specific list of trigger types
// TODO: should allow more then just host as env var to be passed

const Version string = "v1.0.4-beta"
const DefaultRepoPath string = "."
const DefaultHost string = "http://localhost:7071"
const DefaultSortKey string = "name"

var DefaultDocumenterType = documenters.RawDocumenter{}.Name()
var DefaultArgs = map[string]string{}

var Documenters map[string]documenters.Documenter = make(map[string]documenters.Documenter, 4)

func initDocumenters() {
	Documenters[documenters.RawDocumenter{}.Name()] = documenters.RawDocumenter{}
	Documenters[documenters.BrunoDocumenter{}.Name()] = documenters.BrunoDocumenter{}
	Documenters[documenters.MarkdownDocumenter{}.Name()] = documenters.MarkdownDocumenter{}
	Documenters[documenters.InsomniaDocumenter{}.Name()] = documenters.InsomniaDocumenter{}
}

// TODO: remove this, why am I still maintaining this
func writeResults(endpoints []data.EndpointMetaData, docType string, outputDir string, logger *logrus.Logger) {
	// TODO: some of these documenters don't document a single request per file, e.g. insomnia
	for _, endpoint := range endpoints {
		if Documenters[docType].Supports(endpoint.TriggerType) {
			// TODO: Function name is a not a primary key (can have duplicates), live with this overwriting duplicates endpoints for now
			var filePath = path.Join(outputDir, endpoint.Name+Documenters[docType].Extension())
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				logger.Error("Error opening file: " + err.Error())
				return
			}
			defer file.Close()

			var serializedRequest, serializationErr = Documenters[docType].SerializeRequest(endpoint)
			if serializationErr != nil {
				logger.Warn(serializationErr.Error())
				continue
			}
			file.WriteString(serializedRequest)
		}
	}
}

func supportedDocumenters() string {
	stringList := ""
	for _, doc := range Documenters {
		stringList += doc.Name() + ", "
	}
	stringList += "all"
	return stringList
}

// This could have been done better, if I used the same naming
func getDefaultArg(arg string) string {
	if _, yes := DefaultArgs["loaded"]; !yes {
		err := godotenv.Load()
		if err != nil {
			fmt.Println("Error loading .env file for default args")
		}

		DefaultArgs["repo"] = os.Getenv("REPO_PATH")
		DefaultArgs["docType"] = os.Getenv("DOC_TYPE")
		DefaultArgs["outputDir"] = os.Getenv("OUTPUT_DIR")
		DefaultArgs["sortKey"] = os.Getenv("SORT_KEY")
	}

	switch arg {
	case "repo":
		if len(DefaultArgs["repo"]) > 0 {
			return DefaultArgs["repo"]
		}
		return DefaultRepoPath
	case "docType":
		if len(DefaultArgs["docType"]) > 0 {
			return DefaultArgs["docType"]
		}
		return DefaultDocumenterType
	case "outputDir":
		if len(DefaultArgs["outputDir"]) > 0 {
			return DefaultArgs["outputDir"]
		}
		return ""
	case "host":
		if len(DefaultArgs["host"]) > 0 {
			return DefaultArgs["host"]
		}
		return DefaultHost
	case "sortKey":
		if len(DefaultArgs["sortKey"]) > 0 {
			return DefaultArgs["sortKey"]
		}
		return DefaultSortKey
	}
	return ""
}

func getCollectionEnvVars(cmd *flag.FlagSet) map[string]string {
	var collectionEnvVars = make(map[string]string)
	var host = cmd.String("host", getDefaultArg("host"), "host string to prepend the http endpoints with")
	collectionEnvVars["host"] = *host

	return collectionEnvVars
}

func getPrefixKey(p string, prefixes map[string]string) string {
	for k := range prefixes {
		if utils.HasParent(p, k) {
			return k
		}
	}
	return ""
}

func run(logger *logrus.Logger) {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	var repo = runCmd.String("repo", getDefaultArg("repo"), "Path to the repo to parse")
	var docType = runCmd.String("docType", getDefaultArg("docType"), "Documenter type to use ("+supportedDocumenters()+")")
	var outputDir = runCmd.String("outputDir", getDefaultArg("outputDir"), "Dir to output documented api files")
	var endpointSortKey = runCmd.String("sort", getDefaultArg("sortKey"), "the field to sort the endpoints by (name, route, triggerType)")
	runCmd.Parse(os.Args[1:])

	var collectionEnvVars map[string]string = getCollectionEnvVars(runCmd)

	logger.Info("Processing repo: '" + *repo + "' with documenter: '" + *docType + "' will output to: '" + *outputDir + "'")

	if len(*outputDir) > 0 {
		if !utils.InitDir(*outputDir, logger) {
			return
		}
	}

	if _, err := os.Stat(*repo); os.IsNotExist(err) {
		logger.Error("Repo does not exist: " + *repo)
		return
	}

	// locate all the cs files in the repo
	entries, err := utils.GetFiles(*repo, []string{".cs"}, false, true, true)
	if err != nil {
		logger.Error("Error reading repo '" + *repo + "': " + err.Error())
		return
	}

	var prefixes = getApiPrefixes(*repo, logger)
	logger.Debug("Found prefixes: " + strconv.Itoa(len(prefixes)) + " in repo: " + *repo)

	// parse the cs files looking for all the endpoints/triggers
	var endpointCount = 0
	var endpoints = []data.EndpointMetaData{}
	// should this be multithreaded?
	for _, entry := range entries {
		for _, endpoint := range parse(entry, logger) {
			var prefixKey = getPrefixKey(entry.Path, prefixes)
			if prefixes[prefixKey] != "" && len(endpoint.Route) > 0 {
				endpoint.Route = path.Join("/", prefixes[prefixKey], endpoint.Route)
			} else if prefixes[prefixKey] == "" && len(endpoint.Route) > 0 {
				logger.Debug("No prefix found for endpoint: " + endpoint.Name + " in file: " + entry.Path)
			}

			endpoints = append(endpoints, endpoint)
			logger.Info("Found endpoint: " + endpoint.String())
			endpointCount++
		}
	}
	logger.Info("Found " + strconv.Itoa(endpointCount) + " endpoints in repo: " + *repo)

	// sort the endpoints
	sort.Slice(endpoints, func(i, j int) bool {
		if strings.EqualFold(*endpointSortKey, "name") {
			return endpoints[i].Name < endpoints[j].Name
		}
		if strings.EqualFold(*endpointSortKey, "route") {
			return endpoints[i].Route < endpoints[j].Route // this may not be correct
		}
		if strings.EqualFold(*endpointSortKey, "triggerType") {
			return endpoints[i].TriggerType < endpoints[j].TriggerType
		}
		return endpoints[i].Name < endpoints[j].Name
	})

	// begin writing out documentation
	if *docType == "all" {
		for _, doc := range Documenters {
			var outDir = path.Join(*outputDir, doc.Name())
			if !utils.InitDir(outDir, logger) {
				continue
			}
			// writeResults(endpoints, doc.Name(), outDir, logger)
			// TODO: pass "separateFiles" as param from user?
			if !Documenters[doc.Name()].SerializeRequests(endpoints, utils.Base(*repo), outDir, false, collectionEnvVars, logger) {
				logger.Error("Error writing results for documenter: " + doc.Name())
			} else {
				logger.Info("Wrote results for documenter '" + doc.Name() + "' to: " + outDir)
			}
		}
		return
	}

	// maybe determine this before we do all that processing
	if _, exists := Documenters[*docType]; !exists {
		logger.Error("Documenter type '" + *docType + "' does not exist")
		return
	}

	// writeResults(endpoints, *docType, *outputDir, logger)
	if !Documenters[*docType].SerializeRequests(endpoints, utils.Base(*repo), *outputDir, true, collectionEnvVars, logger) {
		logger.Error("Error writing results for documenter: " + Documenters[*docType].Name())
	} else {
		logger.Info("Wrote results for documenter: " + Documenters[*docType].Name() + " to: " + *outputDir)
	}
}

func main() {
	logFile, logger := utils.SetupLogger("combined.log")

	if logFile != nil {
		defer logFile.Close()
	}

	logger.Info("Starting documentApi version: " + Version)
	initDocumenters()

	run(logger)

	logger.Info("Finished documentApi version: " + Version)
}
