package documenters

import (
	"documentApi/data"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

type RawDocumenter struct{}

func (r RawDocumenter) SerializeRequest(endpoint data.EndpointMetaData) (string, error) {
	jsonString, err := json.MarshalIndent(endpoint, "", "    ")
	if err != nil {
		return "", fmt.Errorf("error serializing endpoint: %s", err.Error())
	}

	return string(jsonString), nil
}

func (r RawDocumenter) Name() string {
	return "raw"
}

func (b RawDocumenter) Extension() string {
	return ".json"
}

func (b RawDocumenter) Supports(triggerType string) bool {
	return true
}

func (rawDocumenter RawDocumenter) SerializeRequests(endpoints []data.EndpointMetaData, collectionName string, outputDir string, separateFiles bool, envVars map[string]string, logger *logrus.Logger) bool {
	// vars is not used in this documenter

	if separateFiles {
		// write out the endpoints to individual files
		for _, endpoint := range endpoints {
			// TODO: Function name is a not a primary key (can have duplicates), live with this overwriting duplicates endpoints for now
			var filePath = path.Join(outputDir, endpoint.Name+rawDocumenter.Extension())
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				logger.Error("RawDocumenter SerializeRequests - Error opening endpoint file: " + err.Error())
				return false
			}
			defer file.Close()

			var serializedRequest, serializationErr = rawDocumenter.SerializeRequest(endpoint)
			if serializationErr != nil {
				logger.Warn(serializationErr.Error())
				continue
			}
			var _, writeErr = file.WriteString(serializedRequest)
			if writeErr != nil {
				logger.Error("RawDocumenter SerializeRequests - Error writing endpoint file: " + writeErr.Error())
				return false
			}
		}
	} else {
		var endpointsString string = "["
		for _, endpoint := range endpoints {
			var serializedRequest, serializationErr = rawDocumenter.SerializeRequest(endpoint)
			if serializationErr != nil {
				logger.Warn(serializationErr.Error())
				continue
			}
			endpointsString += fmt.Sprintf("%s,\n", serializedRequest)
		}
		endpointsString = endpointsString[0:len(endpointsString)-2] + "]"
		var filePath = path.Join(outputDir, collectionName+rawDocumenter.Extension())
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			logger.Error("RawDocumenter SerializeRequests - Error opening collection file: " + err.Error())
			return false
		}
		defer file.Close()

		var _, writeErr = file.WriteString(endpointsString)
		if writeErr != nil {
			logger.Error("RawDocumenter SerializeRequests - Error writing collection file: " + writeErr.Error())
			return false
		}
	}

	return true
}
