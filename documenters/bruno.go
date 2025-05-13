package documenters

import (
	"documentApi/data"
	"documentApi/utils"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

type BrunoDocumenter struct{}

// TODO: path parameters that are not immediately preceded by a slash are not handled well by bruno (probably also not insomnia)

var sequence = 1                                           // "static" var to keep track of the sequence number
var pathVarRegex = regexp.MustCompile(`{([a-zA-Z0-9_]+)}`) // TODO: this is duplicated in core.go, should be moved to a common place.

// TODO: handle errors better
func (b BrunoDocumenter) SerializeRequest(endpoint data.EndpointMetaData) string {
	if endpoint.TriggerType != data.TriggerType["Http"] {
		return fmt.Sprintf("Endpoint %s is not an HTTP trigger", endpoint.Name)
	}

	// Implement the serialization logic for Bruno
	var meta = data.BrunoMeta{
		Name: endpoint.Name,
		Type: "http",
		Seq:  sequence,
	}

	var request = data.BrunoRequest{
		URL:  endpoint.Route,
		Body: "json",    // assuming json for now...should probably check if body is empty before adding this field
		Auth: "inherit", // TODO: you probably have to do this yourself
	}

	metaJson, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error serializing meta: %s", err.Error())
	}
	var metaString = strings.ReplaceAll(strings.ReplaceAll(string(metaJson), "\"", ""), ",", "")

	requestJson, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error serializing request: %s", err.Error())
	}
	var requestString = strings.ReplaceAll(strings.ReplaceAll(string(requestJson), "\"", ""), ",", "")

	// ignore body for now
	var bodyString = ""
	// var bodyString = "body:" + request.Body + "{}"

	var pathParamsString = ""
	if len(endpoint.PathParameters) > 0 {
		pathParamsString += "params:path {\n"
		for _, param := range endpoint.PathParameters {
			pathParamsString += fmt.Sprintf("  %s: \n", param)
		}
		pathParamsString += "}"
	}

	sequence++
	// TODO: avoid adding new lines if portions don't exist
	// Only using the first request method, this would probably need to be serialized n times to handle all methods
	// return fmt.Sprintf("meta %s\n\n%s %s\n\nbody:%s {}", metaString, endpoint.Methods[0], requestString, request.Body)
	return fmt.Sprintf("meta %s\n\n%s %s\n\n%s\n\n%s", metaString, endpoint.Methods[0], requestString, pathParamsString, bodyString)
}

func (b BrunoDocumenter) Name() string {
	return "bruno"
}

func (b BrunoDocumenter) Extension() string {
	return ".bru"
}

func (b BrunoDocumenter) Supports(triggerType string) bool {
	return triggerType == data.TriggerType["Http"]
}

func (b BrunoDocumenter) SerializeRequests(endpoints []data.EndpointMetaData, collectionName string, outputDir string, separateFiles bool, logger *logrus.Logger) bool {
	// separateFiles is a no-op for bruno, it expects each endpoint to be in a separate file

	// write out the endpoints to individual files
	// TODO: sort endpoints
	for _, endpoint := range endpoints {
		if b.Supports(endpoint.TriggerType) {
			// TODO: Function name is a not a primary key (can have duplicates), live with this overwriting duplicates endpoints for now
			// consider adding an id to file name to avoid overwriting?
			var filePath = path.Join(outputDir, endpoint.Name+b.Extension())
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				logger.Error("BrunoDocumenter SerializeRequests - Error opening endpoint file: " + err.Error())
				return false
			}
			defer file.Close()

			endpoint.Route = path.Join("{{host}}", replacePathVars(endpoint.Route))
			var _, writeErr = file.WriteString(b.SerializeRequest(endpoint))
			if writeErr != nil {
				logger.Error("BrunoDocumenter SerializeRequests - Error writing endpoint file: " + writeErr.Error())
				return false
			}
		}
	}

	// create the collection file
	var brunoCollectionFile = path.Join(outputDir, "bruno.json")
	file, err := os.OpenFile(brunoCollectionFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("BrunoDocumenter SerializeRequests - Error opening bruno collection file: " + err.Error())
		return false
	}
	defer file.Close()

	var brunoCollection = data.BrunoCollectionDefinition{
		Version: "1",
		Name:    collectionName,
		Type:    "collection",
		Ignore:  []string{"node_modules", ".git"},
	}

	err = json.NewEncoder(file).Encode(brunoCollection)
	if err != nil {
		logger.Error("BrunoDocumenter SerializeRequests - Error writing bruno collection metadata: " + err.Error())
		return false
	}

	// any failing operations after this won't cause the whole serialization to fail, but will cause the environment file to be missing

	// create environment file
	// TODO: this will probably need user customization
	var brunoEnvFile = path.Join(outputDir, "environments", "local.bru")
	if utils.InitDir(path.Join(outputDir, "environments"), logger) {
		file, err := os.OpenFile(brunoEnvFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			logger.Warn("BrunoDocumenter SerializeRequests - Error creating bruno environment file: " + err.Error())
			return true
		}
		defer file.Close()

		var _, writeErr = file.WriteString("vars {\n\t host: http://localhost:7071/\n}") // TODO: get host from cmd params/env file
		if writeErr != nil {
			logger.Warn("BrunoDocumenter SerializeRequests - Error writing bruno environment file: " + writeErr.Error())
			return true
		}
	} else {
		logger.Warn("BrunoDocumenter SerializeRequests - Error creating bruno environment directory")
	}

	return true
}

// TODO: move this to a utils file
func replacePathVars(route string) string {
	// TODO: use pathParameters to replace the path variables in the route
	var results = pathVarRegex.FindAllStringSubmatch(route, -1)

	// no variables
	if results == nil {
		return route
	}

	for _, result := range results {
		route = strings.ReplaceAll(route, result[0], ":"+result[1])
	}

	return route
}
