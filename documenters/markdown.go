package documenters

import (
	"documentApi/data"
	"documentApi/utils"
	"fmt"
	"math"
	"os"
	"path"
	"regexp"
	"strings"

	"slices"

	"github.com/sirupsen/logrus"
)

type MarkdownDocumenter struct{}

// isSeparatorRow checks if a row contains only separator characters (dashes)
func isSeparatorRow(fields []string) bool {
	separatorPattern := regexp.MustCompile(`^ ?[-]+ ?$`)
	for _, field := range fields {
		if !separatorPattern.MatchString(strings.TrimSpace(field)) {
			return false
		}
	}
	return true
}

// adapted from https://github.com/christking246/utils/blob/main/services/Formatter.js#L15
// formatMarkdownTable formats a markdown table with proper alignment
func formatMarkdownTable(str string) string {
	rows := strings.Split(strings.TrimSpace(str), "\n")

	// Parse each row into cells
	var parsedRows [][]string
	for _, row := range rows {
		// Remove leading and trailing pipes, then split by pipe
		trimmed := strings.Trim(row, "|")
		cells := strings.Split(trimmed, "|")
		for i := range cells {
			cells[i] = strings.TrimSpace(cells[i])
		}
		parsedRows = append(parsedRows, cells)
	}

	numColumns := len(parsedRows[0])

	// Calculate maximum width for each column
	columnWidths := make([]int, numColumns)
	var separatorRows []int

	for rowIndex, row := range parsedRows {
		if isSeparatorRow(row) {
			separatorRows = append(separatorRows, rowIndex)
			continue // skip width calculation for separator rows
		}

		for colIndex, cell := range row {
			if colIndex < len(columnWidths) {
				columnWidths[colIndex] = int(math.Max(float64(columnWidths[colIndex]), float64(len(cell)+2)))
			}
		}
	}

	// Rebuild the table with proper alignment
	var formattedRows []string
	for i, row := range parsedRows {
		var dataCells []string
		wasSeparator := slices.Contains(separatorRows, i)
		for index, cell := range row {
			if index >= len(columnWidths) {
				continue
			}

			var padChar, padFront string
			if wasSeparator {
				padChar = "-"
				padFront = ""
			} else {
				padChar = " "
				padFront = " "
			}

			paddedCell := utils.PadRight(padFront+cell, columnWidths[index], padChar)
			dataCells = append(dataCells, paddedCell)
		}
		formattedRows = append(formattedRows, "|"+strings.Join(dataCells, "|")+"|")
	}

	return strings.Join(formattedRows, "\n")
}

func (m MarkdownDocumenter) Extension() string {
	return ".md"
}

func (m MarkdownDocumenter) Name() string {
	return "markdown"
}

func (m MarkdownDocumenter) SerializeRequest(endpoint data.EndpointMetaData) (string, error) {
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s |", endpoint.Name, strings.ToUpper(strings.Join(endpoint.Methods, ", ")), endpoint.Route, strings.Join(endpoint.Authentication, ", "), endpoint.TriggerType, strings.ReplaceAll(endpoint.Interval, "*", "\\*"), endpoint.Description, endpoint.FilePath), nil
}

func (m MarkdownDocumenter) SerializeRequests(endpoints []data.EndpointMetaData, collectionName string, outputDir string, separateFiles bool, vars map[string]string, logger *logrus.Logger) bool {
	// separateFiles is a no-op for markdown, it does not make sense to write a table column per file
	// vars is not used in this documenter

	var markDownString string = "| Function Name | Methods | Route | Authentication | TriggerType | Interval | Description | File Path |\n"
	markDownString += "|--------|--------|--------|--------|--------|--------|--------|--------|\n"
	for _, endpoint := range endpoints {
		var serializedRequest, serializationErr = m.SerializeRequest(endpoint)
		if serializationErr != nil {
			logger.Warn(serializationErr.Error())
			continue
		}
		markDownString += fmt.Sprintf("%s\n", serializedRequest)
	}

	// Format the markdown table for proper alignment
	markDownString = formatMarkdownTable(markDownString) // probably inefficient to do this after building the entire string, instead of doing it while building
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
