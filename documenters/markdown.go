package documenters

import (
	"documentApi/data"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

type MarkdownDocumenter struct{}

func (m MarkdownDocumenter) Extension() string {
	return ".md"
}

func (m MarkdownDocumenter) Name() string {
	return "markdown"
}

func (m MarkdownDocumenter) SerializeRequest(endpoint data.EndpointMetaData) (string, error) {
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", endpoint.Name, strings.Join(endpoint.Methods, ", "), endpoint.Route, strings.Join(endpoint.Authentication, ", "), endpoint.TriggerType, strings.ReplaceAll(endpoint.Interval, "*", "\\*"), endpoint.Description), nil
}

func (m MarkdownDocumenter) SerializeRequests(endpoints []data.EndpointMetaData, collectionName string, outputDir string, separateFiles bool, vars map[string]string, logger *logrus.Logger) bool {
	// separateFiles is a no-op for markdown, it does not make sense to write a table column per file
	// vars is not used in this documenter

	var markDownString string = "| Function Name | Methods | Route | Authentication | TriggerType | Interval | Description |\n"
	markDownString += "| ---------- | ---------- | ---------- | ---------- | ---------- | ---------- | ---------- |\n"
	for _, endpoint := range endpoints {
		var serializedRequest, serializationErr = m.SerializeRequest(endpoint)
		if serializationErr != nil {
			logger.Warn(serializationErr.Error())
			continue
		}
		markDownString += fmt.Sprintf("%s\n", serializedRequest)
	}
	var filePath = path.Join(outputDir, collectionName+m.Extension())
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("MarkdownDocumenter SerializeRequests - Error opening output markdown file: " + err.Error())
		return false
	}
	defer file.Close()

	var _, writeErr = file.WriteString(markDownString)
	if writeErr != nil {
		logger.Error("MarkdownDocumenter SerializeRequests - Error writing markdown file: " + writeErr.Error())
		return false
	}

	return true
}

func (m MarkdownDocumenter) Supports(string) bool {
	return true
}
