package splitter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/akyrey/openapi-splitter/internal/orderedmap"
	"github.com/akyrey/openapi-splitter/internal/util"
	"github.com/akyrey/openapi-splitter/internal/writer"
)

// preservedTopLevel lists fields that are copied directly into the main file
// without being referenced (i.e. not split into separate files).
var preservedTopLevel = []string{
	"openapi", "info", "servers", "tags", "security", "externalDocs", "jsonSchemaDialect",
}

// defaultFieldOrder is the default order of root-level fields in the output file.
var defaultFieldOrder = []string{
	"openapi", "info", "externalDocs", "tags", "paths", "components",
}

// CreateMainFile generates the root openapi.{ext} file with $refs pointing to
// all the split component files.
func CreateMainFile(ctx *Context) error {
	doc := ctx.Doc
	mainDoc := make(map[string]interface{})

	// Copy top-level fields that are kept verbatim
	for _, field := range preservedTopLevel {
		if v, ok := doc[field]; ok {
			mainDoc[field] = v
		}
	}

	// Build paths references
	if paths, ok := doc["paths"].(map[string]interface{}); ok && len(paths) > 0 {
		pathsRef := make(map[string]interface{}, len(paths))
		for pathURL := range paths {
			fileName := util.NormalizePathForFileName(pathURL)
			pathsRef[pathURL] = map[string]interface{}{
				"$ref": "./paths/" + fileName + ctx.Extension,
			}
		}
		mainDoc["paths"] = pathsRef
	}

	// Build components references
	components := make(map[string]interface{})
	compKeys := []struct {
		key string
		dir string
	}{
		{"schemas", "schemas"},
		{"parameters", "parameters"},
		{"responses", "responses"},
		{"requestBodies", "requestBodies"},
		{"headers", "headers"},
		{"securitySchemes", "securitySchemes"},
		{"links", "links"},
		{"callbacks", "callbacks"},
		{"pathItems", "pathItems"},
		{"examples", "examples"},
	}

	if rawComponents, ok := doc["components"].(map[string]interface{}); ok {
		for _, ck := range compKeys {
			if section, ok := rawComponents[ck.key].(map[string]interface{}); ok && len(section) > 0 {
				if ctx.NoIndex {
					// Reference each component file directly
					direct := make(map[string]interface{}, len(section))
					for name := range section {
						direct[name] = map[string]interface{}{
							"$ref": "./" + ck.dir + "/" + name + ctx.Extension,
						}
					}
					components[ck.key] = direct
				} else {
					components[ck.key] = map[string]interface{}{
						"$ref": "./" + ck.dir + "/_index" + ctx.Extension,
					}
				}
			}
		}
	}
	if len(components) > 0 {
		mainDoc["components"] = components
	}

	// Webhooks references (OpenAPI 3.1 top-level field)
	if webhooks, ok := doc["webhooks"].(map[string]interface{}); ok && len(webhooks) > 0 {
		webhooksRef := make(map[string]interface{}, len(webhooks))
		for name := range webhooks {
			fileName := util.NormalizePathForFileName(name)
			webhooksRef[name] = map[string]interface{}{
				"$ref": "./webhooks/" + fileName + ctx.Extension,
			}
		}
		mainDoc["webhooks"] = webhooksRef
	}

	// Determine effective field order
	fieldOrder := defaultFieldOrder
	if len(ctx.FieldOrder) > 0 {
		fieldOrder = ctx.FieldOrder
	}

	// Build an ordered map: first the fields from fieldOrder, then remaining
	// fields in alphabetical order.
	ordered := orderedmap.NewOrderedMap()
	inOrder := make(map[string]bool, len(fieldOrder))
	for _, key := range fieldOrder {
		inOrder[key] = true
		if v, ok := mainDoc[key]; ok {
			ordered.Set(key, v)
		}
	}
	remaining := make([]string, 0, len(mainDoc))
	for key := range mainDoc {
		if !inOrder[key] {
			remaining = append(remaining, key)
		}
	}
	sort.Strings(remaining)
	for _, key := range remaining {
		ordered.Set(key, mainDoc[key])
	}

	outPath := filepath.Join(ctx.OutputDir, "openapi"+ctx.Extension)
	if err := writer.WriteFile(outPath, ordered, ctx.Format, ctx.IndentSize); err != nil {
		return fmt.Errorf("writing main file: %w", err)
	}
	ctx.Logf("wrote main file: %s", outPath)

	return nil
}

// versionPrefix returns "3.0" or "3.1" from a full version string like "3.1.0".
func versionPrefix(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}
