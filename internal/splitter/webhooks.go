package splitter

import (
	"fmt"
	"path/filepath"

	"github.com/akyrey/openapi-splitter/internal/util"
	"github.com/akyrey/openapi-splitter/internal/writer"
)

// SplitWebhooks extracts each webhook from doc["webhooks"] into its own file.
// Webhooks are a top-level field in OpenAPI 3.1.
func SplitWebhooks(ctx *Context) error {
	webhooksRaw, ok := ctx.Doc["webhooks"]
	if !ok {
		ctx.Log("no webhooks found in document")
		return nil
	}

	webhooks, ok := webhooksRaw.(map[string]interface{})
	if !ok || len(webhooks) == 0 {
		ctx.Log("webhooks field is empty or not a map")
		return nil
	}

	ctx.Logf("found %d webhooks to split", len(webhooks))

	for name, item := range webhooks {
		fileName := util.NormalizePathForFileName(name)
		outPath := filepath.Join(ctx.OutputDir, "webhooks", fileName+ctx.Extension)

		rewritten := RewriteRefs(item, ctx.Extension, 1)
		if err := writer.WriteFile(outPath, rewritten, ctx.Format, ctx.IndentSize); err != nil {
			return fmt.Errorf("writing webhook %s: %w", name, err)
		}
		ctx.Logf("wrote %s", outPath)
	}

	return nil
}
