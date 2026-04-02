package splitter_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/akyrey/openapi-splitter/internal/parser"
	"github.com/akyrey/openapi-splitter/internal/splitter"
	"gopkg.in/yaml.v3"
)

// testdataDir returns the absolute path to the project's testdata directory.
func testdataDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../testdata")
	if err != nil {
		t.Fatalf("could not resolve testdata dir: %v", err)
	}
	return dir
}

// splitWith is a convenience helper: parse + split into a temp dir.
func splitWith(t *testing.T, fixture string, opts splitter.Options) string {
	t.Helper()
	outDir := t.TempDir()
	opts.OutputDir = outDir

	doc, err := parser.Parse(fixture, opts.Debug)
	if err != nil {
		t.Fatalf("Parse(%s): %v", fixture, err)
	}
	if err := splitter.Split(doc, opts); err != nil {
		t.Fatalf("Split: %v", err)
	}
	return outDir
}

// assertExists asserts that the given path exists.
func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file/dir to exist: %s", path)
	}
}

// assertNotExists asserts that the given path does not exist.
func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected file/dir NOT to exist: %s", path)
	}
}

// readJSON reads and decodes a JSON file into map[string]interface{}.
func readJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readJSON(%s): %v", path, err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("JSON unmarshal(%s): %v", path, err)
	}
	return m
}

// readYAML reads and decodes a YAML file into map[string]interface{}.
func readYAML(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readYAML(%s): %v", path, err)
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatalf("YAML unmarshal(%s): %v", path, err)
	}
	return m
}

// ─── Test 1: Split JSON OpenAPI 3.0 spec ───────────────────────────────────

func TestSplit_JSON_30(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.0.json")
	out := splitWith(t, fixture, splitter.Options{})

	// Root file
	assertExists(t, filepath.Join(out, "openapi.json"))

	// Directory structure
	for _, dir := range []string{"schemas", "paths", "parameters", "responses"} {
		assertExists(t, filepath.Join(out, dir))
	}

	// Schema files
	assertExists(t, filepath.Join(out, "schemas", "_index.json"))
	assertExists(t, filepath.Join(out, "schemas", "Pet.json"))
	assertExists(t, filepath.Join(out, "schemas", "Pets.json"))
	assertExists(t, filepath.Join(out, "schemas", "Error.json"))

	// Path files
	assertExists(t, filepath.Join(out, "paths", "pets.json"))
	assertExists(t, filepath.Join(out, "paths", "pets_petId.json"))

	// Parameter files
	assertExists(t, filepath.Join(out, "parameters", "_index.json"))
	assertExists(t, filepath.Join(out, "parameters", "limitParam.json"))
	assertExists(t, filepath.Join(out, "parameters", "petId.json"))

	// Response files
	assertExists(t, filepath.Join(out, "responses", "_index.json"))
	assertExists(t, filepath.Join(out, "responses", "Error.json"))

	// Verify main file has correct references
	main := readJSON(t, filepath.Join(out, "openapi.json"))
	info := main["info"].(map[string]interface{})
	if info["title"] != "Swagger Petstore" {
		t.Errorf("info.title = %v, want 'Swagger Petstore'", info["title"])
	}

	// Check path $ref
	paths := main["paths"].(map[string]interface{})
	petsRef := paths["/pets"].(map[string]interface{})["$ref"].(string)
	if petsRef != "./paths/pets.json" {
		t.Errorf("paths[/pets].$ref = %q, want ./paths/pets.json", petsRef)
	}

	// Check components $ref
	components := main["components"].(map[string]interface{})
	schemasRef := components["schemas"].(map[string]interface{})["$ref"].(string)
	if schemasRef != "./schemas/_index.json" {
		t.Errorf("components.schemas.$ref = %q, want ./schemas/_index.json", schemasRef)
	}

	// Verify $refs inside path files are rewritten
	petsFile := readJSON(t, filepath.Join(out, "paths", "pets.json"))
	getOp := petsFile["get"].(map[string]interface{})
	params := getOp["parameters"].([]interface{})
	paramRef := params[0].(map[string]interface{})["$ref"].(string)
	if paramRef != "../parameters/limitParam.json" {
		t.Errorf("paths/pets.json get.parameters[0].$ref = %q, want ../parameters/limitParam.json", paramRef)
	}
}

// ─── Test 2: Split YAML OpenAPI 3.0 spec ───────────────────────────────────

func TestSplit_YAML_30(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.0.yaml")
	out := splitWith(t, fixture, splitter.Options{})

	// Root file should be YAML
	assertExists(t, filepath.Join(out, "openapi.yaml"))
	assertNotExists(t, filepath.Join(out, "openapi.json"))

	// Schema files in YAML
	assertExists(t, filepath.Join(out, "schemas", "_index.yaml"))
	assertExists(t, filepath.Join(out, "schemas", "Pet.yaml"))

	// Verify content of Pet schema
	pet := readYAML(t, filepath.Join(out, "schemas", "Pet.yaml"))
	if pet["type"] != "object" {
		t.Errorf("Pet.type = %v, want object", pet["type"])
	}
}

// ─── Test 3: JSON → YAML format conversion ─────────────────────────────────

func TestSplit_JSON_to_YAML(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.0.json")
	out := splitWith(t, fixture, splitter.Options{Format: "yaml"})

	assertExists(t, filepath.Join(out, "openapi.yaml"))
	assertNotExists(t, filepath.Join(out, "openapi.json"))
	assertExists(t, filepath.Join(out, "schemas", "_index.yaml"))
	assertExists(t, filepath.Join(out, "schemas", "Pet.yaml"))
}

// ─── Test 4: YAML → JSON format conversion ─────────────────────────────────

func TestSplit_YAML_to_JSON(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.0.yaml")
	out := splitWith(t, fixture, splitter.Options{Format: "json"})

	assertExists(t, filepath.Join(out, "openapi.json"))
	assertNotExists(t, filepath.Join(out, "openapi.yaml"))
	assertExists(t, filepath.Join(out, "schemas", "_index.json"))

	main := readJSON(t, filepath.Join(out, "openapi.json"))
	info := main["info"].(map[string]interface{})
	if info["title"] != "Swagger Petstore" {
		t.Errorf("title after format conversion = %v", info["title"])
	}
}

// ─── Test 5: Tags preserved inline (JSON) ───────────────────────────────────

func TestSplit_Tags_JSON(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-with-tags.json")
	out := splitWith(t, fixture, splitter.Options{})

	// No tags directory should be created
	assertNotExists(t, filepath.Join(out, "tags"))

	// Tags must be preserved inline in the main file
	main := readJSON(t, filepath.Join(out, "openapi.json"))
	tagsRaw, ok := main["tags"].([]interface{})
	if !ok {
		t.Fatalf("tags should be an inline array, got %T", main["tags"])
	}
	if len(tagsRaw) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tagsRaw))
	}

	// Verify one tag's content
	petsTag, ok := tagsRaw[0].(map[string]interface{})
	if !ok {
		t.Fatal("first tag is not a map")
	}
	if petsTag["name"] != "pets" {
		t.Errorf("pets tag name = %v", petsTag["name"])
	}
	if petsTag["description"] != "Everything about your Pets" {
		t.Errorf("pets tag description = %v", petsTag["description"])
	}
}

// ─── Test 6: Tags preserved inline (YAML output) ────────────────────────────

func TestSplit_Tags_YAML(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-with-tags.json")
	out := splitWith(t, fixture, splitter.Options{Format: "yaml"})

	// No tags directory should be created
	assertNotExists(t, filepath.Join(out, "tags"))

	// Tags must be preserved inline in the main file
	main := readYAML(t, filepath.Join(out, "openapi.yaml"))
	tagsRaw, ok := main["tags"].([]interface{})
	if !ok {
		t.Fatalf("tags should be an inline array, got %T", main["tags"])
	}
	if len(tagsRaw) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tagsRaw))
	}
	petsTag, ok := tagsRaw[0].(map[string]interface{})
	if !ok {
		t.Fatal("first tag is not a map")
	}
	if petsTag["name"] != "pets" {
		t.Errorf("pets tag name (YAML) = %v", petsTag["name"])
	}
}

// ─── Test 7: Split OpenAPI 3.1 spec ─────────────────────────────────────────

func TestSplit_JSON_31(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.1.json")
	out := splitWith(t, fixture, splitter.Options{})

	// Standard components
	assertExists(t, filepath.Join(out, "openapi.json"))
	assertExists(t, filepath.Join(out, "schemas", "Pet.json"))
	assertExists(t, filepath.Join(out, "parameters", "limitParam.json"))
	assertExists(t, filepath.Join(out, "responses", "Error.json"))

	// Extended components (3.1)
	assertExists(t, filepath.Join(out, "requestBodies", "PetBody.json"))
	assertExists(t, filepath.Join(out, "headers", "X-Rate-Limit.json"))
	assertExists(t, filepath.Join(out, "examples", "PetExample.json"))
	assertExists(t, filepath.Join(out, "securitySchemes", "BearerAuth.json"))

	// Webhooks (3.1 top-level)
	assertExists(t, filepath.Join(out, "webhooks", "newPet.json"))
	assertExists(t, filepath.Join(out, "webhooks", "petDeleted.json"))

	// Tags are preserved inline (no tags directory)
	assertNotExists(t, filepath.Join(out, "tags"))
}

// ─── Test 8: 3.1 fields preserved in main file ──────────────────────────────

func TestSplit_31_FieldsPreserved(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.1.json")
	out := splitWith(t, fixture, splitter.Options{})

	main := readJSON(t, filepath.Join(out, "openapi.json"))

	// openapi version
	if main["openapi"] != "3.1.0" {
		t.Errorf("openapi = %v, want 3.1.0", main["openapi"])
	}

	// jsonSchemaDialect should be preserved
	if main["jsonSchemaDialect"] == nil {
		t.Error("jsonSchemaDialect not preserved in main file")
	}

	// webhooks should be referenced
	webhooks, ok := main["webhooks"].(map[string]interface{})
	if !ok {
		t.Fatal("webhooks not present in main file")
	}
	newPetRef := webhooks["newPet"].(map[string]interface{})["$ref"].(string)
	if newPetRef != "./webhooks/newPet.json" {
		t.Errorf("webhooks.newPet.$ref = %q, want ./webhooks/newPet.json", newPetRef)
	}
}

// ─── Test 9: Extended component $refs rewritten correctly ───────────────────

func TestSplit_31_RefRewriting(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.1.json")
	out := splitWith(t, fixture, splitter.Options{})

	// In the paths/pets.json file, the POST requestBody should point to requestBodies/
	petsPath := readJSON(t, filepath.Join(out, "paths", "pets.json"))
	post := petsPath["post"].(map[string]interface{})
	rbRef := post["requestBody"].(map[string]interface{})["$ref"].(string)
	if rbRef != "../requestBodies/PetBody.json" {
		t.Errorf("paths/pets.json post.requestBody.$ref = %q, want ../requestBodies/PetBody.json", rbRef)
	}
}

// ─── Test 11: --no-index with OpenAPI 3.0 ───────────────────────────────────

func TestSplit_NoIndex_JSON_30(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.0.json")
	out := splitWith(t, fixture, splitter.Options{NoIndex: true})

	// Individual component files must still exist
	assertExists(t, filepath.Join(out, "schemas", "Pet.json"))
	assertExists(t, filepath.Join(out, "schemas", "Pets.json"))
	assertExists(t, filepath.Join(out, "schemas", "Error.json"))
	assertExists(t, filepath.Join(out, "parameters", "limitParam.json"))
	assertExists(t, filepath.Join(out, "parameters", "petId.json"))
	assertExists(t, filepath.Join(out, "responses", "Error.json"))

	// _index files must NOT exist
	assertNotExists(t, filepath.Join(out, "schemas", "_index.json"))
	assertNotExists(t, filepath.Join(out, "parameters", "_index.json"))
	assertNotExists(t, filepath.Join(out, "responses", "_index.json"))

	// Root file must reference individual component files directly
	main := readJSON(t, filepath.Join(out, "openapi.json"))
	components := main["components"].(map[string]interface{})

	schemas := components["schemas"].(map[string]interface{})
	petRef := schemas["Pet"].(map[string]interface{})["$ref"].(string)
	if petRef != "./schemas/Pet.json" {
		t.Errorf("components.schemas.Pet.$ref = %q, want ./schemas/Pet.json", petRef)
	}
	errorRef := schemas["Error"].(map[string]interface{})["$ref"].(string)
	if errorRef != "./schemas/Error.json" {
		t.Errorf("components.schemas.Error.$ref = %q, want ./schemas/Error.json", errorRef)
	}

	params := components["parameters"].(map[string]interface{})
	limitRef := params["limitParam"].(map[string]interface{})["$ref"].(string)
	if limitRef != "./parameters/limitParam.json" {
		t.Errorf("components.parameters.limitParam.$ref = %q, want ./parameters/limitParam.json", limitRef)
	}

	// Paths still reference individual files (unchanged by --no-index)
	paths := main["paths"].(map[string]interface{})
	petsRef := paths["/pets"].(map[string]interface{})["$ref"].(string)
	if petsRef != "./paths/pets.json" {
		t.Errorf("paths[/pets].$ref = %q, want ./paths/pets.json", petsRef)
	}

	// $refs inside path files still resolve correctly (depth=1)
	petsFile := readJSON(t, filepath.Join(out, "paths", "pets.json"))
	getOp := petsFile["get"].(map[string]interface{})
	params2 := getOp["parameters"].([]interface{})
	paramRef := params2[0].(map[string]interface{})["$ref"].(string)
	if paramRef != "../parameters/limitParam.json" {
		t.Errorf("paths/pets.json get.parameters[0].$ref = %q, want ../parameters/limitParam.json", paramRef)
	}
}

// ─── Test 13: --no-index with OpenAPI 3.1 ───────────────────────────────────

func TestSplit_NoIndex_JSON_31(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.1.json")
	out := splitWith(t, fixture, splitter.Options{NoIndex: true})

	// Individual files must still exist
	assertExists(t, filepath.Join(out, "schemas", "Pet.json"))
	assertExists(t, filepath.Join(out, "requestBodies", "PetBody.json"))
	assertExists(t, filepath.Join(out, "headers", "X-Rate-Limit.json"))
	assertExists(t, filepath.Join(out, "examples", "PetExample.json"))
	assertExists(t, filepath.Join(out, "securitySchemes", "BearerAuth.json"))

	// _index files must NOT exist
	assertNotExists(t, filepath.Join(out, "schemas", "_index.json"))
	assertNotExists(t, filepath.Join(out, "requestBodies", "_index.json"))
	assertNotExists(t, filepath.Join(out, "headers", "_index.json"))

	// Root file references individual component files directly
	main := readJSON(t, filepath.Join(out, "openapi.json"))
	components := main["components"].(map[string]interface{})

	schemas := components["schemas"].(map[string]interface{})
	petRef := schemas["Pet"].(map[string]interface{})["$ref"].(string)
	if petRef != "./schemas/Pet.json" {
		t.Errorf("components.schemas.Pet.$ref = %q, want ./schemas/Pet.json", petRef)
	}

	requestBodies := components["requestBodies"].(map[string]interface{})
	rbRef := requestBodies["PetBody"].(map[string]interface{})["$ref"].(string)
	if rbRef != "./requestBodies/PetBody.json" {
		t.Errorf("components.requestBodies.PetBody.$ref = %q, want ./requestBodies/PetBody.json", rbRef)
	}

	// Webhooks are unaffected by --no-index (they never used an index file)
	webhooks := main["webhooks"].(map[string]interface{})
	newPetRef := webhooks["newPet"].(map[string]interface{})["$ref"].(string)
	if newPetRef != "./webhooks/newPet.json" {
		t.Errorf("webhooks.newPet.$ref = %q, want ./webhooks/newPet.json", newPetRef)
	}
}

func TestSplit_FieldOrder_JSON_Default(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-with-tags.json")
	outDir := splitWith(t, fixture, splitter.Options{})

	data, err := os.ReadFile(filepath.Join(outDir, "openapi.json"))
	if err != nil {
		t.Fatalf("read openapi.json: %v", err)
	}

	keys := jsonTopLevelKeys(t, data)
	// Default order: openapi, info, externalDocs, tags, paths, components.
	// petstore-with-tags has: openapi, info, tags, paths, components (no externalDocs).
	assertKeysBefore(t, keys, "openapi", "info")
	assertKeysBefore(t, keys, "info", "tags")
	assertKeysBefore(t, keys, "tags", "paths")
	assertKeysBefore(t, keys, "paths", "components")
}

func TestSplit_FieldOrder_YAML_Default(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-with-tags.json")
	outDir := splitWith(t, fixture, splitter.Options{Format: "yaml"})

	data, err := os.ReadFile(filepath.Join(outDir, "openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}

	keys := yamlTopLevelKeys(t, data)
	assertKeysBefore(t, keys, "openapi", "info")
	assertKeysBefore(t, keys, "info", "tags")
	assertKeysBefore(t, keys, "tags", "paths")
	assertKeysBefore(t, keys, "paths", "components")
}

func TestSplit_FieldOrder_Custom(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-with-tags.json")
	customOrder := []string{"paths", "components", "openapi", "info"}
	outDir := splitWith(t, fixture, splitter.Options{FieldOrder: customOrder})

	data, err := os.ReadFile(filepath.Join(outDir, "openapi.json"))
	if err != nil {
		t.Fatalf("read openapi.json: %v", err)
	}

	keys := jsonTopLevelKeys(t, data)
	assertKeysBefore(t, keys, "paths", "components")
	assertKeysBefore(t, keys, "components", "openapi")
	assertKeysBefore(t, keys, "openapi", "info")
	// tags is not in custom order so it appears after the ordered fields
	assertKeysBefore(t, keys, "info", "tags")
}

// jsonTopLevelKeys returns the top-level keys of a JSON object in order.
func jsonTopLevelKeys(t *testing.T, data []byte) []string {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		t.Fatalf("expected '{' token: %v", err)
	}
	var keys []string
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			t.Fatalf("token error: %v", err)
		}
		key, ok := tok.(string)
		if !ok {
			t.Fatalf("expected string key, got %T", tok)
		}
		keys = append(keys, key)
		var val interface{}
		if err := dec.Decode(&val); err != nil {
			t.Fatalf("decode value: %v", err)
		}
	}
	return keys
}

// yamlTopLevelKeys returns the top-level keys of a YAML mapping in order.
func yamlTopLevelKeys(t *testing.T, data []byte) []string {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		t.Fatalf("expected DocumentNode with content")
	}
	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		t.Fatalf("expected MappingNode, got kind %d", mapping.Kind)
	}
	var keys []string
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		keys = append(keys, mapping.Content[i].Value)
	}
	return keys
}

// assertKeysBefore asserts that key `a` appears before key `b` in keys.
func assertKeysBefore(t *testing.T, keys []string, a, b string) {
	t.Helper()
	posA, posB := -1, -1
	for i, k := range keys {
		if k == a {
			posA = i
		}
		if k == b {
			posB = i
		}
	}
	if posA == -1 {
		t.Errorf("key %q not found in %v", a, keys)
		return
	}
	if posB == -1 {
		t.Errorf("key %q not found in %v", b, keys)
		return
	}
	if posA >= posB {
		t.Errorf("expected %q (pos %d) before %q (pos %d) in %v", a, posA, b, posB, keys)
	}
}

func TestSplit_OutputDirIsClean(t *testing.T) {
	fixture := filepath.Join(testdataDir(t), "petstore-3.0.json")

	outDir := t.TempDir()
	// Write a stale file that should be removed
	staleFile := filepath.Join(outDir, "stale.txt")
	_ = os.WriteFile(staleFile, []byte("old content"), 0o644)

	doc, err := parser.Parse(fixture, false)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := splitter.Split(doc, splitter.Options{OutputDir: outDir}); err != nil {
		t.Fatalf("Split: %v", err)
	}

	// Stale file should no longer exist
	assertNotExists(t, staleFile)
}
