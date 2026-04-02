// Package splitter provides the context shared across all splitting operations.
package splitter

import "fmt"

// Options holds the user-provided options for a split operation.
type Options struct {
	InputFile string
	OutputDir string
	Format    string // "json" or "yaml"; empty means infer from input
	Debug     bool
	// NoIndex disables creation of _index.{ext} files. When true, the root
	// openapi file references individual component/tag files directly.
	NoIndex bool
	// FieldOrder defines the order of root-level fields in the main output file.
	// Nil means use the default order.
	FieldOrder []string
	// IndentSize is the number of spaces used for indentation in output files.
	IndentSize int
}

// Context holds all shared state for a single split operation.
type Context struct {
	// Doc is the raw parsed OpenAPI document as a generic map.
	Doc map[string]interface{}
	// OutputDir is the absolute path to the output directory.
	OutputDir string
	// Format is the resolved output format: "json" or "yaml".
	Format string
	// Extension is the file extension including the dot: ".json" or ".yaml".
	Extension string
	// Debug enables verbose logging.
	Debug bool
	// Version is the detected OpenAPI version prefix: "3.0" or "3.1".
	Version string
	// NoIndex disables creation of _index.{ext} files.
	NoIndex bool
	// FieldOrder defines the order of root-level fields in the main output file.
	// Nil means use the default order.
	FieldOrder []string
	// IndentSize is the number of spaces used for indentation in output files.
	IndentSize int
}

// Log prints a debug message if debug mode is enabled.
func (c *Context) Log(msg string) {
	if c.Debug {
		fmt.Println("[DEBUG]", msg)
	}
}

// Logf prints a formatted debug message if debug mode is enabled.
func (c *Context) Logf(format string, args ...interface{}) {
	if c.Debug {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}
