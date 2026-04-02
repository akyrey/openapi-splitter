package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/akyrey/openapi-splitter/internal/parser"
	"github.com/akyrey/openapi-splitter/internal/splitter"
	editorconfig "github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

const defaultIndentSize = 2

var (
	outputDir       string
	format          string
	debug           bool
	noIndex         bool
	fieldOrder      []string
	indent          int
	useEditorconfig bool
)

var rootCmd = &cobra.Command{
	Use:     "openapi-splitter <file>",
	Short:   "Split an OpenAPI specification into multiple files",
	Long:    `A CLI tool to split a single OpenAPI 3.0/3.1 specification file (JSON or YAML) into multiple organized files with proper $ref references.`,
	Version: version,
	Args:    cobra.ExactArgs(1),
	RunE:    run,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "./openapi-split", "Output directory")
	rootCmd.Flags().StringVarP(&format, "format", "f", "", "Output format: json or yaml (default: inferred from input)")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
	rootCmd.Flags().BoolVarP(&noIndex, "no-index", "n", false, "Disable _index file creation; root file references individual component files directly")
	rootCmd.Flags().StringSliceVar(&fieldOrder, "field-order", nil, "Comma-separated root field order (default: openapi,info,externalDocs,tags,paths,components)")
	rootCmd.Flags().IntVarP(&indent, "indent", "i", 0, "Number of spaces for indentation (default: 2, or from .editorconfig if --editorconfig is set)")
	rootCmd.Flags().BoolVarP(&useEditorconfig, "editorconfig", "e", false, "Read indentation settings from .editorconfig in the current directory")
}

func run(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Validate format flag if provided
	if format != "" && format != "json" && format != "yaml" {
		return fmt.Errorf("invalid format %q: must be 'json' or 'yaml'", format)
	}

	// Check input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", inputFile)
	}

	// Resolve indent size: --indent flag > editorconfig > default
	indentSize, err := resolveIndentSize(cmd, format, inputFile)
	if err != nil {
		return err
	}

	// Parse the input file
	doc, err := parser.Parse(inputFile, debug)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI file: %w", err)
	}

	// Build splitter options
	opts := splitter.Options{
		InputFile:  inputFile,
		OutputDir:  outputDir,
		Format:     format,
		Debug:      debug,
		NoIndex:    noIndex,
		FieldOrder: fieldOrder,
		IndentSize: indentSize,
	}

	// Run the splitter
	if err := splitter.Split(doc, opts); err != nil {
		return fmt.Errorf("failed to split OpenAPI specification: %w", err)
	}

	return nil
}

// resolveIndentSize determines the indent size to use. Priority:
// 1. Explicit --indent flag
// 2. .editorconfig in CWD (when --editorconfig is set)
// 3. Default (2)
func resolveIndentSize(cmd *cobra.Command, outputFormat string, inputFile string) (int, error) {
	// If --indent was explicitly provided, use it directly
	if cmd.Flags().Changed("indent") {
		if indent <= 0 {
			return 0, fmt.Errorf("--indent must be a positive integer")
		}
		return indent, nil
	}

	// Try .editorconfig if requested
	if useEditorconfig {
		size, found, err := indentFromEditorconfig(outputFormat, inputFile)
		if err != nil {
			return 0, fmt.Errorf("reading .editorconfig: %w", err)
		}
		if found {
			return size, nil
		}
	}

	return defaultIndentSize, nil
}

// indentFromEditorconfig looks up indent_size for the given format from
// .editorconfig starting at CWD. Returns (size, found, error).
func indentFromEditorconfig(outputFormat string, inputFile string) (int, bool, error) {
	// Determine extension to look up
	ext := outputFormat
	if ext == "" {
		// Infer from input file extension
		switch filepath.Ext(inputFile) {
		case ".json":
			ext = "json"
		default:
			ext = "yaml"
		}
	}

	// Use a dummy filename in CWD so editorconfig searches upward from there
	cwd, err := os.Getwd()
	if err != nil {
		return 0, false, err
	}
	dummyFile := filepath.Join(cwd, "openapi."+ext)

	def, err := editorconfig.GetDefinitionForFilename(dummyFile)
	if err != nil {
		// No .editorconfig found is not an error
		return 0, false, nil
	}

	if def.IndentSize == "" || def.IndentSize == "tab" {
		return 0, false, nil
	}

	var size int
	if _, err := fmt.Sscanf(def.IndentSize, "%d", &size); err != nil || size <= 0 {
		return 0, false, nil
	}

	return size, true, nil
}
