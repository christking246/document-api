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
		route = strings.ReplaceAll(route, result[0], ":"+result[1])
	}

	return route
}

// ExtractPathVars extracts the path variables from a route string.
// It returns a slice of strings containing the variable names.
// For example, for the route "/api/{id}/details", it will return ["id"].
func ExtractPathVars(route string) []string {
	var params []string = []string{}

	var matches = pathVarRegex.FindAllStringSubmatch(route, -1)
	if len(matches) > 0 {
		for _, match := range matches {
			params = append(params, match[1])
		}
	}

	return params
}
