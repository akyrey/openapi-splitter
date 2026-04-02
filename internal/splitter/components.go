// Package splitter contains helpers for splitting component maps.
package splitter

import (
	"fmt"
	"path/filepath"

	"github.com/akyrey/openapi-splitter/internal/writer"
)

// splitComponentMap extracts a component section (e.g. schemas, parameters) from
// the document's components object, writing each entry as an individual file and
// generating an _index file that references them all.
//
// componentKey is the key inside doc["components"] (e.g. "schemas").
// dirName is the output subdirectory name (e.g. "schemas").
func splitComponentMap(ctx *Context, componentKey, dirName string) error {
	components, ok := ctx.Doc["components"].(map[string]interface{})
	if !ok {
		ctx.Logf("no components object found, skipping %s", componentKey)
		return nil
	}

	section, ok := components[componentKey].(map[string]interface{})
	if !ok || len(section) == 0 {
		ctx.Logf("no %s found in components", componentKey)
		return nil
	}

	ctx.Logf("found %d %s to split", len(section), componentKey)

	indexContent := make(map[string]interface{}, len(section))

	for name, item := range section {
		outPath := filepath.Join(ctx.OutputDir, dirName, name+ctx.Extension)
		// Rewrite internal $refs inside this component (depth=1: one level deep)
		rewritten := RewriteRefs(item, ctx.Extension, 1)
		if err := writer.WriteFile(outPath, rewritten, ctx.Format, ctx.IndentSize); err != nil {
			return fmt.Errorf("writing %s/%s: %w", dirName, name, err)
		}
		ctx.Logf("wrote %s", outPath)
		indexContent[name] = map[string]interface{}{
			"$ref": "./" + name + ctx.Extension,
		}
	}

	if ctx.NoIndex {
		return nil
	}

	// Write the _index file
	indexPath := filepath.Join(ctx.OutputDir, dirName, "_index"+ctx.Extension)
	if err := writer.WriteFile(indexPath, indexContent, ctx.Format, ctx.IndentSize); err != nil {
		return fmt.Errorf("writing %s/_index: %w", dirName, err)
	}
	ctx.Logf("wrote %s", indexPath)

	return nil
}
