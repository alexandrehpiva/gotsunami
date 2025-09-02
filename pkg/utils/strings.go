package utils

import (
	"strings"
)

// ExpandTemplate expands template variables in a string
func ExpandTemplate(template string, variables map[string]string) string {
	result := template

	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ContainsAny checks if a string contains any of the given substrings
func ContainsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// ContainsAll checks if a string contains all of the given substrings
func ContainsAll(s string, substrings []string) bool {
	for _, substr := range substrings {
		if !strings.Contains(s, substr) {
			return false
		}
	}
	return true
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// JoinNonEmpty joins non-empty strings with a separator
func JoinNonEmpty(separator string, stringSlice ...string) string {
	var nonEmpty []string
	for _, s := range stringSlice {
		if !IsEmpty(s) {
			nonEmpty = append(nonEmpty, s)
		}
	}
	return strings.Join(nonEmpty, separator)
}
