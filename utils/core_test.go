package utils

import (
	"testing"
)

func Test_ReplacePathVars_ReturnsPathWithVariableIdentifiersForBrunoAndInsomnia(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "No variables",
			path:     "/api/details/all",
			expected: "/api/details/all",
		},
		{
			name:     "Single variable end",
			path:     "/api/details/{id}",
			expected: "/api/details/:id",
		},
		{
			name:     "Single variable middle",
			path:     "/api/{user}/score",
			expected: "/api/:user/score",
		},
		{
			name:     "Multiple variables",
			path:     "/api/{user}/details/{course}",
			expected: "/api/:user/details/:course",
		},
		{
			name:     "Variable with partial value (front)",
			path:     "/api/{user}/details/course:{course}",
			expected: "/api/:user/details/:course",
		},
		{
			name:     "Variable with partial value (back)",
			path:     "/api/{user}/details/{course}:course",
			expected: "/api/:user/details/:course",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ReplacePathVars(test.path)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func Test_ExtractPathVars_ReturnsMapContainingAllPathVars(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected map[string]string
	}{
		{
			name:     "No variables",
			path:     "/api/details/all",
			expected: map[string]string{},
		},
		{
			name:     "Single variable end",
			path:     "/api/details/{id}",
			expected: map[string]string{"id": ""},
		},
		{
			name:     "Single variable middle",
			path:     "/api/{user}/score",
			expected: map[string]string{"user": ""},
		},
		{
			name:     "Multiple variables",
			path:     "/api/{user}/details/{course}",
			expected: map[string]string{"user": "", "course": ""},
		},
		{
			name:     "Variable with partial value",
			path:     "/api/{user}/details/course:{course}",
			expected: map[string]string{"user": "", "course": "course:"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractPathVars(test.path)
			AssertMapEqual(t, test.expected, result)
		})
	}
}

func Test_GetPartialPathParamValue_ReturnsNonVariablePortionOfPathParam(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		param    string
		expected string
	}{
		{
			name:     "No constant in path",
			path:     "/api/details/{id}",
			param:    "id",
			expected: "",
		},
		{
			name:     "Parameter at end",
			path:     "/api/details/id:{id}",
			param:    "id",
			expected: "id:",
		},
		{
			name:     "Parameter in the middle",
			path:     "/api/user:{user}/score",
			param:    "user",
			expected: "user:",
		},
		{
			name:     "Parameter at end with constant at the end",
			path:     "/api/{user}/details/{course}:course",
			param:    "course",
			expected: ":course",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getPartialPathParamValue(test.path, test.param)
			if result != test.expected {
				t.Errorf("expected %s, got '%s'", test.expected, result)
			}
		})
	}
}
