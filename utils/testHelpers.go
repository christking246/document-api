package utils

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

// mapToStringArray converts a map to a slice of strings in the format "key=value".
func mapToStringArray(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k, v := range m {
		result = append(result, k+"="+v)
	}
	return result
}

func AssertContains(t *testing.T, list []string, str string) bool {
	if slices.Contains(list, str) {
		return true
	}
	t.Errorf("%s", t.Name()+" - Expected list: ["+strings.Join(list, ", ")+"] to contain: \""+str+"\", but it did not.")
	return false
}

func AssertNotContains(t *testing.T, list []string, str string) bool {
	if slices.Contains(list, str) {
		t.Errorf("%s", t.Name()+" - Expected list: ["+strings.Join(list, ", ")+"] to not contain: \""+str+"\", but it did.")
		return false
	}
	return true
}

func AssertStringEqual(t *testing.T, expected string, actual string) bool {
	if expected != actual {
		t.Errorf("%s", t.Name()+" - Expected: \""+expected+"\", but got: \""+actual+"\"")
		return false
	}
	return true
}

// intentionally not using interface{}
func AssertEqual(t *testing.T, expected int, actual int) bool {
	if expected != actual {
		t.Errorf("%s", t.Name()+" - Expected: \""+strconv.Itoa(expected)+"\", but got: \""+strconv.Itoa(actual)+"\"")
		return false
	}
	return true
}

func AssertMin(t *testing.T, min int, value int) bool {
	if value < min {
		t.Errorf("%s", t.Name()+" - Expected: \""+strconv.Itoa(value)+"\", to be at least: \""+strconv.Itoa(min)+"\"")
		return false
	}
	return true
}

func AssertMax(t *testing.T, max int, value int) bool {
	if value > max {
		t.Errorf("%s", t.Name()+" - Expected: \""+strconv.Itoa(value)+"\", to be at most: \""+strconv.Itoa(max)+"\"")
		return false
	}
	return true
}

// list order must match
func AssertSliceEqual(t *testing.T, expected []string, actual []string) bool {
	if len(expected) != len(actual) {
		t.Errorf("%s", t.Name()+" - Expected: ["+strings.Join(expected, ", ")+"] but got: ["+strings.Join(actual, ", ")+"]")
		return false
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("%s", t.Name()+" - Expected: ["+strings.Join(expected, ", ")+"] but got: ["+strings.Join(actual, ", ")+"]")
			return false
		}
	}
	return true
}

func AssertMapEqual(t *testing.T, expected map[string]string, actual map[string]string) bool {
	if len(expected) != len(actual) {
		t.Errorf("%s", t.Name()+" - Expected: ["+strings.Join(mapToStringArray(expected), ", ")+"] but got: ["+strings.Join(mapToStringArray(actual), ", ")+"]")
		return false
	}
	for k, v := range expected {
		if actual[k] != v {
			t.Errorf("%s", t.Name()+" - Expected: ["+strings.Join(mapToStringArray(expected), ", ")+"] but got: ["+strings.Join(mapToStringArray(actual), ", ")+"]")
			return false
		}
	}
	return true
}

func AssertMapContains(t *testing.T, m map[string]string, k string) bool {
	if _, ok := m[k]; !ok {
		t.Errorf("%s", t.Name()+" - Expected map: ["+strings.Join(mapToStringArray(m), ", ")+"] to contain key: \""+k+"\", but it did not.")
		return false
	}
	return true
}
