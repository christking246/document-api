package utils

import (
	"regexp"
	"strings"
)

var pathVarRegex = regexp.MustCompile(`{([a-zA-Z0-9_]+)}`)

// ReplacePathVars replaces the path variables in a route string with a colon followed by the variable name.
// For example, for the route "/api/{id}/details", it will return "/api/:id/details".
// This is the format used by the bruno and insomnia
func ReplacePathVars(route string) string {
	// TODO: use pathParameters to replace the path variables in the route?
	var results = pathVarRegex.FindAllStringSubmatch(route, -1)

	// no variables
	if results == nil {
		return route
	}

	for _, result := range results {
		var partialValue = getPartialPathParamValue(route, result[1])
		// TODO: write that replace by regex function
		route = strings.ReplaceAll(route, partialValue+result[0], ":"+result[1])
		route = strings.ReplaceAll(route, result[0]+partialValue, ":"+result[1])
	}

	return route
}

// ExtractPathVars extracts the path variables from a route string.
// It returns a map of the param names to their partial values (if exist).
// For example, for the route "/api/{id}/details/user:{user}", it will return ["id" -> "", "user" -> "user:"].
func ExtractPathVars(route string) map[string]string {
	var params map[string]string = make(map[string]string)

	var matches = pathVarRegex.FindAllStringSubmatch(route, -1)
	if len(matches) > 0 {
		for _, match := range matches {
			var partialValue = getPartialPathParamValue(route, match[1])
			params[match[1]] = partialValue
		}
	}

	return params
}

// GetPartialPathParamValue returns the non-variable portion of a path parameter in a route string.
// For example, for the route "/api/details/id:{id}", it will return "id:".
func getPartialPathParamValue(route string, paramName string) string {
	var re = regexp.MustCompile(`/([^/]+)?\{` + paramName + `\}([^/]+)?/?`)

	var matches = re.FindStringSubmatch(route)

	if len(matches) > 0 {
		// only support grabbing for the front OR back of the variable portion of param
		if matches[1] != "" {
			return matches[1]
		}
		if matches[2] != "" {
			return matches[2]
		}
	}
	return ""
}
