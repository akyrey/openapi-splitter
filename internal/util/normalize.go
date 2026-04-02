// Package util provides shared utility functions.
package util

import (
	"regexp"
	"strings"
)

var (
	reNonAlphanumeric = regexp.MustCompile(`[^a-zA-Z0-9]`)
	reMultiUnderscore = regexp.MustCompile(`_{2,}`)
)

// NormalizePathForFileName converts an OpenAPI path string into a safe filename.
// For example: "/pets/{petId}" -> "pets_petId_"
func NormalizePathForFileName(pathURL string) string {
	// Remove leading slash
	s := strings.TrimPrefix(pathURL, "/")
	// Replace all non-alphanumeric characters with underscores
	s = reNonAlphanumeric.ReplaceAllString(s, "_")
	// Collapse consecutive underscores
	s = reMultiUnderscore.ReplaceAllString(s, "_")
	// Remove trailing underscore if present
	s = strings.TrimSuffix(s, "_")
	return s
}
