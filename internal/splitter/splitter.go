// Package splitter orchestrates the splitting of an OpenAPI document.
package splitter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/akyrey/openapi-splitter/internal/parser"
)

// Split is the main entry point. It takes a parsed document and options and
// produces the split directory structure.
func Split(doc *parser.ParsedDoc, opts Options) error {
	// Resolve output directory to an absolute path
	outputDir, err := filepath.Abs(opts.OutputDir)
	if err != nil {
		return fmt.Errorf("could not resolve output directory: %w", err)
	}

	// Determine output format
	format := opts.Format
	if format == "" {
		format = doc.InputFormat
	}

	extension := "." + format

	// Determine indent size; default to 2 if not specified
	indentSize := opts.IndentSize
	if indentSize <= 0 {
		indentSize = 2
	}

	// Build context
	ctx := &Context{
		Doc:        doc.Raw,
		OutputDir:  outputDir,
		Format:     format,
		Extension:  extension,
		Debug:      opts.Debug,
		Version:    versionPrefix(doc.Version),
		NoIndex:    opts.NoIndex,
		FieldOrder: opts.FieldOrder,
		IndentSize: indentSize,
	}

	fmt.Printf("Reading OpenAPI %s specification...\n", doc.Version)
	fmt.Printf("Creating output directory structure in: %s...\n", outputDir)

	// Remove and recreate the output directory for a clean slate
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("could not clean output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}

	// Pre-create all expected subdirectories
	subdirs := []string{"paths", "schemas", "parameters", "responses",
		"requestBodies", "headers", "securitySchemes", "links",
		"callbacks", "pathItems", "examples", "webhooks"}
	for _, sub := range subdirs {
		if err := os.MkdirAll(filepath.Join(outputDir, sub), 0o755); err != nil {
			return fmt.Errorf("could not create subdirectory %s: %w", sub, err)
		}
	}

	fmt.Println("Splitting OpenAPI specification...")

	// Run all splitters in order
	steps := []struct {
		name string
		fn   func(*Context) error
	}{
		{"schemas", SplitSchemas},
		{"paths", SplitPaths},
		{"parameters", SplitParameters},
		{"responses", SplitResponses},
		{"requestBodies", SplitRequestBodies},
		{"headers", SplitHeaders},
		{"securitySchemes", SplitSecuritySchemes},
		{"links", SplitLinks},
		{"callbacks", SplitCallbacks},
		{"pathItems", SplitPathItems},
		{"examples", SplitExamples},
		{"webhooks", SplitWebhooks},
	}

	for _, step := range steps {
		ctx.Logf("running splitter: %s", step.name)
		if err := step.fn(ctx); err != nil {
			return fmt.Errorf("%s splitter failed: %w", step.name, err)
		}
	}

	// Generate the main openapi file
	if err := CreateMainFile(ctx); err != nil {
		return fmt.Errorf("creating main file: %w", err)
	}

	// Clean up empty subdirectories
	if err := removeEmptyDirs(outputDir, subdirs); err != nil {
		ctx.Logf("warning: could not remove empty dirs: %v", err)
	}

	fmt.Printf("\nOpenAPI specification successfully split into %s\n", outputDir)
	return nil
}

// removeEmptyDirs removes subdirectories that contain no files.
func removeEmptyDirs(outputDir string, subdirs []string) error {
	for _, sub := range subdirs {
		dir := filepath.Join(outputDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			// Directory might not exist if splitter skipped it
			continue
		}
		if len(entries) == 0 {
			if err := os.Remove(dir); err != nil {
				return err
			}
		}
	}
	return nil
}
