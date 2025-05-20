package utils

import "testing"

func Test_ReplacePathVars_ReturnsCorrectStringForBasePath(t *testing.T) {
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

func Test_ExtractPathVars_ReturnsCorrectStringForBasePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "No variables",
			path:     "/api/details/all",
			expected: []string{},
		},
		{
			name:     "Single variable end",
			path:     "/api/details/{id}",
			expected: []string{"id"},
		},
		{
			name:     "Single variable middle",
			path:     "/api/{user}/score",
			expected: []string{"user"},
		},
		{
			name:     "Multiple variables",
			path:     "/api/{user}/details/{course}",
			expected: []string{"user", "course"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ExtractPathVars(test.path)
			AssertSliceEqual(t, test.expected, result)
		})
	}
}
