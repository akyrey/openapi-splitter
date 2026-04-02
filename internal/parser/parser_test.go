package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/akyrey/openapi-splitter/internal/parser"
)

// testdataDir returns the absolute path to the project's testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	// Walk up from the package directory to find testdata/
	dir, err := filepath.Abs("../../testdata")
	if err != nil {
		t.Fatalf("could not resolve testdata dir: %v", err)
	}
	return dir
}

func TestParse_JSON_30(t *testing.T) {
	doc, err := parser.Parse(filepath.Join(testdataDir(t), "petstore-3.0.json"), false)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if doc.Version != "3.0.3" {
		t.Errorf("Version = %q, want 3.0.3", doc.Version)
	}
	if doc.InputFormat != "json" {
		t.Errorf("InputFormat = %q, want json", doc.InputFormat)
	}
	if doc.Raw["info"] == nil {
		t.Error("missing 'info' field in parsed doc")
	}
}

func TestParse_YAML_30(t *testing.T) {
	doc, err := parser.Parse(filepath.Join(testdataDir(t), "petstore-3.0.yaml"), false)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if doc.Version != "3.0.3" {
		t.Errorf("Version = %q, want 3.0.3", doc.Version)
	}
	if doc.InputFormat != "yaml" {
		t.Errorf("InputFormat = %q, want yaml", doc.InputFormat)
	}
}

func TestParse_JSON_31(t *testing.T) {
	doc, err := parser.Parse(filepath.Join(testdataDir(t), "petstore-3.1.json"), false)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if doc.Version != "3.1.0" {
		t.Errorf("Version = %q, want 3.1.0", doc.Version)
	}
	// Verify 3.1-specific field is present in raw doc
	if doc.Raw["jsonSchemaDialect"] == nil {
		t.Error("missing 'jsonSchemaDialect' field in 3.1 doc")
	}
	if doc.Raw["webhooks"] == nil {
		t.Error("missing 'webhooks' field in 3.1 doc")
	}
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := parser.Parse("/nonexistent/file.json", false)
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.json")
	os.WriteFile(f, []byte("{invalid json"), 0o644)
	_, err := parser.Parse(f, false)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParse_MissingOpenapiField(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "no_version.json")
	os.WriteFile(f, []byte(`{"info": {"title": "Test", "version": "1.0"}}`), 0o644)
	_, err := parser.Parse(f, false)
	if err == nil {
		t.Error("expected error for missing 'openapi' field, got nil")
	}
}

func TestParse_UnsupportedVersion(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "v2.json")
	os.WriteFile(f, []byte(`{"openapi": "2.0", "info": {"title": "T", "version": "1"}}`), 0o644)
	_, err := parser.Parse(f, false)
	if err == nil {
		t.Error("expected error for Swagger 2.0, got nil")
	}
}
