package writer_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/akyrey/openapi-splitter/internal/writer"
	"gopkg.in/yaml.v3"
)

func TestWriteFile_JSON(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.json")

	content := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}

	if err := writer.WriteFile(outPath, content, "json", 2); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("key = %v, want %v", result["key"], "value")
	}
}

func TestWriteFile_YAML(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.yaml")

	content := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}

	if err := writer.WriteFile(outPath, content, "yaml", 2); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("YAML unmarshal error: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("key = %v, want %v", result["key"], "value")
	}
}

func TestWriteFile_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "deeply", "nested", "dir", "file.json")

	if err := writer.WriteFile(outPath, map[string]interface{}{"x": 1}, "json", 2); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Errorf("file not created at %s", outPath)
	}
}

func TestWriteFile_JSONPrettyPrinted(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "pretty.json")

	content := map[string]interface{}{"a": "b"}
	if err := writer.WriteFile(outPath, content, "json", 2); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	data, _ := os.ReadFile(outPath)
	// Pretty-printed JSON should contain newlines
	found := false
	for _, b := range data {
		if b == '\n' {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected pretty-printed JSON with newlines, got: %s", data)
	}
}

func TestWriteFile_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	err := writer.WriteFile(filepath.Join(dir, "f.txt"), map[string]interface{}{}, "xml", 2)
	if err == nil {
		t.Error("expected error for invalid format, got nil")
	}
}

func TestWriteFile_JSONIndentSize(t *testing.T) {
	nested := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "value",
		},
	}

	tests := []struct {
		indentSize int
		wantPrefix string
	}{
		{2, "  "},
		{4, "    "},
	}

	for _, tc := range tests {
		t.Run("indent"+strings.Repeat(" ", tc.indentSize), func(t *testing.T) {
			dir := t.TempDir()
			outPath := filepath.Join(dir, "test.json")
			if err := writer.WriteFile(outPath, nested, "json", tc.indentSize); err != nil {
				t.Fatalf("WriteFile error: %v", err)
			}
			data, _ := os.ReadFile(outPath)
			content := string(data)
			if !strings.Contains(content, "\n"+tc.wantPrefix) {
				t.Errorf("indent size %d: expected lines indented with %q, got:\n%s", tc.indentSize, tc.wantPrefix, content)
			}
		})
	}
}

func TestWriteFile_YAMLIndentSize(t *testing.T) {
	nested := map[string]interface{}{
		"outer": map[string]interface{}{
			"inner": "value",
		},
	}

	tests := []struct {
		indentSize int
		wantPrefix string
	}{
		{2, "  inner"},
		{4, "    inner"},
	}

	for _, tc := range tests {
		t.Run("indent"+strings.Repeat(" ", tc.indentSize), func(t *testing.T) {
			dir := t.TempDir()
			outPath := filepath.Join(dir, "test.yaml")
			if err := writer.WriteFile(outPath, nested, "yaml", tc.indentSize); err != nil {
				t.Fatalf("WriteFile error: %v", err)
			}
			data, _ := os.ReadFile(outPath)
			content := string(data)
			if !strings.Contains(content, tc.wantPrefix) {
				t.Errorf("indent size %d: expected %q in output, got:\n%s", tc.indentSize, tc.wantPrefix, content)
			}
		})
	}
}
