package utils

import "testing"

func Test_Base_ReturnsCorrectStringForBasePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Empty path",
			path:     "",
			expected: ".",
		},
		{
			name:     "Absolute path",
			path:     "/home/user/project",
			expected: "project",
		},
		{
			name:     "Windows path",
			path:     "C:\\Users\\User\\Documents",
			expected: "Documents",
		},
		{
			name:     "Relative path",
			path:     "src/utils",
			expected: "utils",
		},
		{
			name:     "Relative windows path",
			path:     "src\\utils",
			expected: "utils",
		},
		{
			name:     "Already base path",
			path:     "run.py",
			expected: "run.py",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Base(test.path)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func Test_Dir_ReturnsCorrectStringForDirPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Empty path",
			path:     "",
			expected: ".",
		},
		{
			name:     "Absolute path",
			path:     "/home/user/project",
			expected: "/home/user",
		},
		{
			name:     "Windows path",
			path:     "C:\\Users\\User\\Documents",
			expected: "C:/Users/User",
		},
		{
			name:     "Relative path",
			path:     "src/utils",
			expected: "src",
		},
		{
			name:     "Relative windows path",
			path:     "src\\utils",
			expected: "src",
		},
		{
			name:     "base path",
			path:     "run.py",
			expected: ".",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Dir(test.path)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func Test_HasParent_ReturnsCorrectBoolean(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		parent   string
		expected bool
	}{
		{
			name:     "Empty path and parent",
			path:     "",
			parent:   "",
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			parent:   "/home/user",
			expected: false,
		},
		{
			name:     "Empty parent",
			path:     "/home/user/project",
			parent:   "",
			expected: false,
		},
		{
			name:     "Path has direct parent",
			path:     "/home/user/project",
			parent:   "/home/user",
			expected: true,
		},
		{
			name:     "Path has indirect parent",
			path:     "/home/user/project/src/utils",
			parent:   "/home/user/project",
			expected: true,
		},
		{
			name:     "Path has direct parent (windows)",
			path:     "C:\\home\\user\\project",
			parent:   "C:\\home\\user",
			expected: true,
		},
		{
			name:     "Path has indirect parent (windows)",
			path:     "C:\\home\\user\\project\\src\\utils",
			parent:   "C:\\home\\user\\project",
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := HasParent(test.path, test.parent)
			if result != test.expected {
				t.Errorf("expected %t, got %t", test.expected, result)
			}
		})
	}
}
