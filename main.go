package main

import (
	"context"
	"documentApi/data"
	"documentApi/documenters"
	"documentApi/utils"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sirupsen/logrus"
)

// TODO: add option to keep old vars (env, path params, etc) from existing collections upon updating
// TODO: add option to create documentation for a specific list of trigger types
// TODO: should allow more then just host as env var to be passed

const Version string = "v1.0.7-beta"
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

func process(params data.RunParams, logger *logrus.Logger) {
	logger.Info("Processing repo: '" + *params.Repo + "' with documenter: '" + *params.DocType + "' will output to: '" + *params.OutputDir + "'")

	if len(*params.OutputDir) > 0 {
		if !utils.InitDir(*params.OutputDir, logger) {
			return
		}
	}

	if _, err := os.Stat(*params.Repo); os.IsNotExist(err) {
		logger.Error("Repo does not exist: " + *params.Repo)
		return
	}

	// locate all the cs files in the repo
	entries, err := utils.GetFiles(*params.Repo, []string{".cs"}, false, true, true)
	if err != nil {
		logger.Error("Error reading repo '" + *params.Repo + "': " + err.Error())
		return
	}

	var prefixes = getApiPrefixes(*params.Repo, logger)
	logger.Debug("Found prefixes: " + strconv.Itoa(len(prefixes)) + " in repo: " + *params.Repo)

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
	logger.Info("Found " + strconv.Itoa(endpointCount) + " endpoints in repo: " + *params.Repo)

	// sort the endpoints
	sort.Slice(endpoints, func(i, j int) bool {
		if strings.EqualFold(*params.EndpointSortKey, "name") {
			return endpoints[i].Name < endpoints[j].Name
		}
		if strings.EqualFold(*params.EndpointSortKey, "route") {
			return endpoints[i].Route < endpoints[j].Route // this may not be correct
		}
		if strings.EqualFold(*params.EndpointSortKey, "triggerType") {
			return endpoints[i].TriggerType < endpoints[j].TriggerType
		}
		return endpoints[i].Name < endpoints[j].Name
	})

	// begin writing out documentation
	if *params.DocType == "all" {
		for _, doc := range Documenters {
			var outDir = path.Join(*params.OutputDir, doc.Name())
			if !utils.InitDir(outDir, logger) {
				continue
			}
			// writeResults(endpoints, doc.Name(), outDir, logger)
			// TODO: pass "separateFiles" as param from user?
			if !Documenters[doc.Name()].SerializeRequests(endpoints, utils.Base(*params.Repo), outDir, false, params.CollectionEnvVars, logger) {
				logger.Error("Error writing results for documenter: " + doc.Name())
			} else {
				logger.Info("Wrote results for documenter '" + doc.Name() + "' to: " + outDir)
			}
		}
		return
	}

	// maybe determine this before we do all that processing
	if _, exists := Documenters[*params.DocType]; !exists {
		logger.Error("Documenter type '" + *params.DocType + "' does not exist")
		return
	}

	// writeResults(endpoints, *docType, *outputDir, logger)
	if !Documenters[*params.DocType].SerializeRequests(endpoints, utils.Base(*params.Repo), *params.OutputDir, true, params.CollectionEnvVars, logger) {
		logger.Error("Error writing results for documenter: " + Documenters[*params.DocType].Name())
	} else {
		logger.Info("Wrote results for documenter: " + Documenters[*params.DocType].Name() + " to: " + *params.OutputDir)
	}
}

func run(logger *logrus.Logger) {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	RunParams := data.RunParams{}
	RunParams.Repo = runCmd.String("repo", getDefaultArg("repo"), "Path to the repo to parse")
	RunParams.DocType = runCmd.String("docType", getDefaultArg("docType"), "Documenter type to use ("+supportedDocumenters()+")")
	RunParams.OutputDir = runCmd.String("outputDir", getDefaultArg("outputDir"), "Dir to output documented api files")
	RunParams.EndpointSortKey = runCmd.String("sort", getDefaultArg("sortKey"), "the field to sort the endpoints by (name, route, triggerType)")
	runCmd.Parse(os.Args[2:])

	RunParams.CollectionEnvVars = getCollectionEnvVars(runCmd)

	process(RunParams, logger)
}

func serve(logger *logrus.Logger) {
	logger.Info("Running as server")

	// Run the server
	var port = "8080"
	if os.Getenv("SERVER_PORT") != "" {
		port = os.Getenv("SERVER_PORT")
	}

	type Empty struct{}

	type VersionOutput struct {
		Version string `json:"version" jsonschema:"the version of the API"`
	}

	mcpRun := func(ctx context.Context, req *mcp.CallToolRequest, input data.RunParams) (*mcp.CallToolResult, Empty, error) {
		if input.OutputDir == nil {
			outputDir := getDefaultArg("outputDir")
			input.OutputDir = &outputDir
		}

		if input.EndpointSortKey == nil {
			endpointSortKey := getDefaultArg("sortKey")
			input.EndpointSortKey = &endpointSortKey
		}

		// TODO: implement function for settings collection env vars, in a unified way between cli and server ... for now set host as the default if unset
		if input.CollectionEnvVars == nil || len(input.CollectionEnvVars) == 0 {
			input.CollectionEnvVars = make(map[string]string)
			input.CollectionEnvVars["host"] = getDefaultArg("host")
		}

		process(input, logger)
		return nil, Empty{}, nil
	}

	mcpPing := func(ctx context.Context, req *mcp.CallToolRequest, input Empty) (*mcp.CallToolResult, VersionOutput, error) {
		return nil, VersionOutput{Version: Version}, nil
	}

	// Create a server with a single tool.
	server := mcp.NewServer(&mcp.Implementation{Name: "documentApi", Version: Version}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "version", Description: "get the version"}, mcpPing)
	mcp.AddTool(server, &mcp.Tool{Name: "document", Description: "generate the api documentation"}, mcpRun)
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	// TODO: allow https
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	httpServer.ListenAndServe()
}

func main() {
	logFile, logger := utils.SetupLogger("combined.log")

	if logFile != nil {
		defer logFile.Close()
	}

	logger.Info("Starting documentApi version: " + Version)
	initDocumenters()

	if len(os.Args) < 2 {
		logger.Error("Missing subcommand")
		return
	}

	switch os.Args[1] {
	case "run":
		run(logger)
		logger.Info("Finished documentApi version: " + Version)
	case "serve":
		serve(logger)
	case "version":
		logger.Info("Version: " + Version)
	default:
		logger.Error("Unknown subcommand: " + os.Args[1])
	}
}
