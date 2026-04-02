// Package splitter contains the $ref rewriting logic.
package splitter

import (
	"strings"
)

// componentRefMap maps internal $ref prefixes to relative output paths.
// The value is the directory name relative to the output root.
var componentRefMap = map[string]string{
	"#/components/schemas/":         "schemas",
	"#/components/parameters/":      "parameters",
	"#/components/responses/":       "responses",
	"#/components/requestBodies/":   "requestBodies",
	"#/components/headers/":         "headers",
	"#/components/securitySchemes/": "securitySchemes",
	"#/components/links/":           "links",
	"#/components/callbacks/":       "callbacks",
	"#/components/pathItems/":       "pathItems",
	"#/components/examples/":        "examples",
}

// RewriteRefs recursively walks obj and rewrites all $ref values that point to
// internal component definitions so they point to the split files instead.
// The depth parameter tracks nesting so the relative path prefix ("../") is
// applied correctly. Files inside a component directory (e.g. schemas/) need
// one level up ("../") to reach a sibling directory.
func RewriteRefs(obj interface{}, ext string, depth int) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			if key == "$ref" {
				if s, ok := val.(string); ok {
					result[key] = rewriteRef(s, ext, depth)
					continue
				}
			}
			result[key] = RewriteRefs(val, ext, depth)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = RewriteRefs(item, ext, depth)
		}
		return result

	default:
		return obj
	}
}

// rewriteRef converts a single $ref string if it matches a known component prefix.
func rewriteRef(ref, ext string, depth int) string {
	for prefix, dir := range componentRefMap {
		if strings.HasPrefix(ref, prefix) {
			name := strings.TrimPrefix(ref, prefix)
			// depth == 1: inside a component file (e.g. paths/foo.json) → ../schemas/X.ext
			// depth == 0: inside the root openapi file → ./schemas/X.ext
			if depth > 0 {
				return "../" + dir + "/" + name + ext
			}
			return "./" + dir + "/" + name + ext
		}
	}
	// Unknown ref — leave unchanged
	return ref
}
