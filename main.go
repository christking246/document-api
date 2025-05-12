package main

import (
	"documentApi/data"
	"documentApi/documenters"
	"documentApi/utils"
	"encoding/json"
	"flag"
	"os"
	"path"
	"strconv"

	"github.com/sirupsen/logrus"
)

// TODO: add option to keep old vars (env, path params, etc) from existing collections upon updating
// TODO: add option to create documentation for a specific list of trigger types
// TODO: allow loading cmds params from a .env file

const Version string = "v1.0.0-alpha"
const DefaultRepoPath string = "."

var DefaultDocumenterType = documenters.RawDocumenter{}.Name()

var Documenters map[string]documenters.Documenter = make(map[string]documenters.Documenter, 2)

func initDocumenters() {
	Documenters[documenters.RawDocumenter{}.Name()] = documenters.RawDocumenter{}
	Documenters[documenters.BrunoDocumenter{}.Name()] = documenters.BrunoDocumenter{}
}

func writeResults(endpoints []data.EndpointMetaData, docType string, outputDir string, logger *logrus.Logger) {
	// TODO: some of these documenters don't document a single request per file, e.g. insomnia
	// TODO: sort endpoints
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

			file.WriteString(Documenters[docType].SerializeRequest(endpoint))
		}
	}
}

func run(logger *logrus.Logger) {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	var repo = runCmd.String("repo", DefaultRepoPath, "Path to the repo to parse")
	var docType = runCmd.String("docType", DefaultDocumenterType, "Documenter type to use (raw, bruno, insomnia, markdown, web, all)") // TODO: only list the active/supported documenters
	var outputDir = runCmd.String("outputDir", "", "Dir to output documented api files")
	runCmd.Parse(os.Args[1:])

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

	// read host json file to get the path prepended to all endpoints
	jsonEntries, err := utils.GetFiles(*repo, []string{".json"}, false, true, true)
	if err != nil {
		logger.Warn("Error reading repo '" + *repo + "' to locate host json file: " + err.Error())
	}

	// TODO: write array filtering function?
	var apiPrefix string = ""
	for _, entry := range jsonEntries {
		if entry.Name == "host.json" {
			hostFileData, err := os.ReadFile(entry.Path)
			if err != nil {
				logger.Warn("Error reading host json file: " + entry.Path + ": " + err.Error())
				return
			}
			var hostData data.ApiMetaData
			if err := json.Unmarshal(hostFileData, &hostData); err != nil {
				logger.Warn("Error parsing host file: " + entry.Path + ": " + err.Error())
				return
			}
			apiPrefix = hostData.Extensions.Http.RoutePrefix
		}
	}

	// parse the cs files looking for all the endpoints/triggers
	var endpointCount = 0
	var endpoints = []data.EndpointMetaData{}
	// should this be multithreaded?
	for _, entry := range entries {
		for _, endpoint := range parse(entry, logger) {
			if len(apiPrefix) > 0 && len(endpoint.Route) > 0 {
				endpoint.Route = path.Join("/", apiPrefix, endpoint.Route)
			}

			endpoints = append(endpoints, endpoint)
			logger.Info("Found endpoint: " + endpoint.String())
			endpointCount++
		}
	}
	logger.Info("Found " + strconv.Itoa(endpointCount) + " endpoints in repo: " + *repo)

	// begin writing out documentation
	if *docType == "all" {
		for _, doc := range Documenters {
			var outDir = path.Join(*outputDir, doc.Name())
			if !utils.InitDir(outDir, logger) {
				continue
			}
			// writeResults(endpoints, doc.Name(), outDir, logger)
			// TODO: pass "separateFiles" as param from user?
			if !Documenters[doc.Name()].SerializeRequests(endpoints, utils.Base(*repo), outDir, false, logger) {
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
	if !Documenters[*docType].SerializeRequests(endpoints, utils.Base(*repo), *outputDir, true, logger) {
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
