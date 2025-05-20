package main

import (
	"documentApi/data"
	"documentApi/utils"
	"os"
	"strings"
	"testing"
)

var _ = os.Setenv("ENV", "test")
var _, testLogger = utils.SetupLogger("test.log")

func readTestFile(filePath string) []string {
	fileData, _ := os.ReadFile(filePath)
	var fileDataString string = string(fileData)
	fileDataString = strings.ReplaceAll(fileDataString, "\r\n", "\n") // In case windows is using \r\n
	return strings.Split(fileDataString, "\n")
}

func Test_parseFunctionHeader_ReturnsCorrectNumberOfSkippedLines(t *testing.T) {
	// Arrange
	codeFileLines := readTestFile("test_assets/one_endpoint.cs")
	codeFileData := strings.Join(codeFileLines, "\n")
	var endpoint data.EndpointMetaData

	// Act
	var skipped = parseFunctionHeader(codeFileData[1440:], &endpoint)

	// Assert
	utils.AssertEqual(t, 2, skipped)
}

func Test_parseFunctionHeader_ReturnsZeroWhenNoLinesSkipped(t *testing.T) {
	// Arrange
	codeFileLines := readTestFile("test_assets/one_endpoint_one_line_header.cs")
	codeFileData := strings.Join(codeFileLines, "\n")
	var endpoint data.EndpointMetaData

	// Act
	var skipped = parseFunctionHeader(codeFileData[470:], &endpoint)

	// Assert
	utils.AssertEqual(t, 0, skipped)
}

func Test_parseFunctionHeader_ReturnsExpectedFunctionMetaData(t *testing.T) {
	// Arrange
	codeFileLines := readTestFile("test_assets/one_endpoint.cs")
	codeFileData := strings.Join(codeFileLines, "\n")
	var endpoint data.EndpointMetaData

	// Act
	parseFunctionHeader(codeFileData[1440:], &endpoint)

	// Assert
	utils.AssertStringEqual(t, data.TriggerType["Http"], endpoint.TriggerType)
	utils.AssertEqual(t, 1, len(endpoint.Methods))
	utils.AssertStringEqual(t, "get", endpoint.Methods[0])
	utils.AssertStringEqual(t, "dashboard-summary/{param}", endpoint.Route)
	utils.AssertEqual(t, 1, len(endpoint.PathParameters))
	utils.AssertStringEqual(t, "param", endpoint.PathParameters[0])

}

func Test_parseFunctionHeader_ReturnsExpectedFunctionMetaDataFromSingleLineHeader(t *testing.T) {
	// Arrange
	codeFileLines := readTestFile("test_assets/one_endpoint_one_line_header.cs")
	codeFileData := strings.Join(codeFileLines, "\n")
	var endpoint data.EndpointMetaData

	// Act
	parseFunctionHeader(codeFileData[470:], &endpoint)

	// Assert
	utils.AssertStringEqual(t, data.TriggerType["Http"], endpoint.TriggerType)
	utils.AssertEqual(t, 1, len(endpoint.Methods))
	utils.AssertStringEqual(t, "get", endpoint.Methods[0])
	utils.AssertStringEqual(t, "dashboard-summary", endpoint.Route)
}

func Test_parse_ReturnsAllExistingHttpTriggers(t *testing.T) {
	// Arrange
	var testFile = data.FileMetaData{
		Name: "http_endpoints_and_helpers.cs",
		Path: "test_assets/http_endpoints_and_helpers.cs",
	}

	// Act
	var endpoints = parse(testFile, testLogger)

	// Assert
	utils.AssertEqual(t, 4, len(endpoints))
}

func Test_parse_ReturnsAllDataOnExistingHttpTriggers(t *testing.T) {
	// Arrange
	var testFile = data.FileMetaData{
		Name: "http_endpoints_and_helpers.cs",
		Path: "test_assets/http_endpoints_and_helpers.cs",
	}

	var expectedEndpoints []data.EndpointMetaData = make([]data.EndpointMetaData, 0, 4)
	expectedEndpoints = append(expectedEndpoints, data.EndpointMetaData{
		Name:           "GetInitialInfoAsync",
		Route:          "sandbox/{moduleId}/info",
		Methods:        []string{"get"},
		PathParameters: []string{"moduleId"},
		TriggerType:    data.TriggerType["Http"],
	})
	expectedEndpoints = append(expectedEndpoints, data.EndpointMetaData{
		Name:           "GetAsync",
		Authentication: []string{"OperationType.Read"},
		Route:          "sandbox/{moduleId}",
		Methods:        []string{"get"},
		PathParameters: []string{"moduleId"},
		TriggerType:    data.TriggerType["Http"],
	})
	expectedEndpoints = append(expectedEndpoints, data.EndpointMetaData{
		Name:           "PreprovisionSandboxAsync",
		Authentication: []string{"DocsToken"},
		Route:          "sandbox/preprovision/{moduleId}",
		Methods:        []string{"post"},
		PathParameters: []string{"moduleId"},
		TriggerType:    data.TriggerType["Http"],
	})
	expectedEndpoints = append(expectedEndpoints, data.EndpointMetaData{
		Name:        "VerifyModules",
		Route:       "sandbox/verify",
		Methods:     []string{"get"},
		TriggerType: data.TriggerType["Http"],
	})

	// Act
	var endpoints = parse(testFile, testLogger)

	// Assert
	for i := range expectedEndpoints {
		utils.AssertStringEqual(t, expectedEndpoints[i].Name, endpoints[i].Name)
		utils.AssertStringEqual(t, expectedEndpoints[i].Route, endpoints[i].Route)
		utils.AssertStringEqual(t, expectedEndpoints[i].TriggerType, endpoints[i].TriggerType)
		utils.AssertEqual(t, len(expectedEndpoints[i].Methods), len(endpoints[i].Methods))
		utils.AssertEqual(t, len(expectedEndpoints[i].PathParameters), len(endpoints[i].PathParameters))
		utils.AssertEqual(t, len(expectedEndpoints[i].Authentication), len(endpoints[i].Authentication))
		for j := range expectedEndpoints[i].Methods {
			utils.AssertStringEqual(t, expectedEndpoints[i].Methods[j], endpoints[i].Methods[j])
		}
		for j := range expectedEndpoints[i].PathParameters {
			utils.AssertStringEqual(t, expectedEndpoints[i].PathParameters[j], endpoints[i].PathParameters[j])
		}
		for j := range expectedEndpoints[i].Authentication {
			utils.AssertStringEqual(t, expectedEndpoints[i].Authentication[j], endpoints[i].Authentication[j])
		}
	}
}

func Test_searchAuthentication_ReturnsExpectedAuthentication(t *testing.T) {
	// Arrange
	tests := []struct {
		name          string
		expectedAuths []string
		authString    string
	}{
		{
			name:          "DocsTokenGroups",
			expectedAuths: []string{"Learn", "Dirt-box", "Config", "Admin", "SG"},
			authString:    "[RequireDocsTokenGroups(\"Learn Dirt-box Config Admin SG\")]",
		},
		{
			name:          "DocsToken Read",
			expectedAuths: []string{"OperationType.Read"},
			authString:    "[RequireDocsToken(OperationType.Read)]",
		},
		{
			name:          "S2SToken",
			expectedAuths: []string{"S2S.SkillLessons"},
			authString:    "[RequireS2SToken(S2S.SkillLessons)]",
		},
		{
			name:          "S2SToken Multiple",
			expectedAuths: []string{"WLW", "S2S.Percentile"},
			authString:    "[RequireS2SToken(\"WLW\", S2S.Percentile)]",
		},
		{
			name:          "DocsToken",
			expectedAuths: []string{"DocsToken"},
			authString:    "[RequireDocsToken]",
		},
		{
			name:          "PlatformToken",
			expectedAuths: []string{"PlatformApiAuth"},
			authString:    "[RequirePlatformApiAuth]",
		},
	}
	var results = make([]data.EndpointMetaData, len(tests))

	// Act && Assert
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			searchAuthentication(test.authString, &results[i])
			utils.AssertSliceEqual(t, test.expectedAuths, results[i].Authentication)
		})
	}
}
