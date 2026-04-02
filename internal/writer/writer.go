// Package writer serialises content to JSON or YAML files.
package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WriteFile serialises content to filePath in the given format ("json" or "yaml").
// It creates all parent directories as needed.
// indentSize controls the number of spaces per indentation level.
func WriteFile(filePath string, content interface{}, format string, indentSize int) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("could not create directory %s: %w", filepath.Dir(filePath), err)
	}

	var data []byte
	var err error

	indent := strings.Repeat(" ", indentSize)

	switch format {
	case "json":
		data, err = json.Marshal(content)
		if err != nil {
			return fmt.Errorf("JSON marshal error for %s: %w", filePath, err)
		}
		var indented bytes.Buffer
		if err = json.Indent(&indented, data, "", indent); err != nil {
			return fmt.Errorf("JSON indent error for %s: %w", filePath, err)
		}
		data = indented.Bytes()
		// Ensure trailing newline
		data = append(data, '\n')
	case "yaml":
		data, err = marshalYAML(content, indentSize)
		if err != nil {
			return fmt.Errorf("YAML marshal error for %s: %w", filePath, err)
		}
	default:
		return fmt.Errorf("unknown format %q: must be 'json' or 'yaml'", format)
	}

	// Write atomically: write to a temp file then rename
	tmp := filePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("could not write temp file %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, filePath); err != nil {
		// Fallback: direct write if rename fails (e.g. cross-device)
		_ = os.Remove(tmp)
		if err2 := os.WriteFile(filePath, data, 0o644); err2 != nil {
			return fmt.Errorf("could not write file %s: %w", filePath, err2)
		}
	}

	return nil
}

// marshalYAML converts content to YAML bytes with the specified indent size.
func marshalYAML(content interface{}, indentSize int) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indentSize)
	if err := enc.Encode(content); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
