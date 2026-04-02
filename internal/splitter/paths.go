package splitter

import (
	"fmt"
	"path/filepath"

	"github.com/akyrey/openapi-splitter/internal/util"
	"github.com/akyrey/openapi-splitter/internal/writer"
)

// SplitPaths extracts each path from doc["paths"] into its own file.
func SplitPaths(ctx *Context) error {
	paths, ok := ctx.Doc["paths"].(map[string]interface{})
	if !ok || len(paths) == 0 {
		ctx.Log("no paths found in document")
		return nil
	}

	ctx.Logf("found %d paths to split", len(paths))

	for pathURL, pathItem := range paths {
		fileName := util.NormalizePathForFileName(pathURL)
		outPath := filepath.Join(ctx.OutputDir, "paths", fileName+ctx.Extension)

		ctx.Logf("processing path %q -> %s", pathURL, fileName+ctx.Extension)

		// Rewrite internal $refs inside this path item (depth=1)
		rewritten := RewriteRefs(pathItem, ctx.Extension, 1)
		if err := writer.WriteFile(outPath, rewritten, ctx.Format, ctx.IndentSize); err != nil {
			return fmt.Errorf("writing path %s: %w", pathURL, err)
		}
		ctx.Logf("wrote %s", outPath)
	}

	return nil
}
