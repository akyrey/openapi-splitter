package splitter_test

import (
	"testing"

	"github.com/akyrey/openapi-splitter/internal/splitter"
)

func TestRewriteRefs_SchemaRef(t *testing.T) {
	input := map[string]interface{}{
		"$ref": "#/components/schemas/Pet",
	}
	got := splitter.RewriteRefs(input, ".json", 1)
	m, ok := got.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}
	want := "../schemas/Pet.json"
	if m["$ref"] != want {
		t.Errorf("got $ref %q, want %q", m["$ref"], want)
	}
}

func TestRewriteRefs_ParameterRef(t *testing.T) {
	input := map[string]interface{}{
		"$ref": "#/components/parameters/limitParam",
	}
	got := splitter.RewriteRefs(input, ".yaml", 1)
	m := got.(map[string]interface{})
	want := "../parameters/limitParam.yaml"
	if m["$ref"] != want {
		t.Errorf("got $ref %q, want %q", m["$ref"], want)
	}
}

func TestRewriteRefs_ResponseRef(t *testing.T) {
	input := map[string]interface{}{
		"$ref": "#/components/responses/Error",
	}
	got := splitter.RewriteRefs(input, ".json", 1)
	m := got.(map[string]interface{})
	want := "../responses/Error.json"
	if m["$ref"] != want {
		t.Errorf("got $ref %q, want %q", m["$ref"], want)
	}
}

func TestRewriteRefs_RequestBodyRef(t *testing.T) {
	input := map[string]interface{}{
		"$ref": "#/components/requestBodies/PetBody",
	}
	got := splitter.RewriteRefs(input, ".json", 1)
	m := got.(map[string]interface{})
	want := "../requestBodies/PetBody.json"
	if m["$ref"] != want {
		t.Errorf("got $ref %q, want %q", m["$ref"], want)
	}
}

func TestRewriteRefs_AllComponentTypes(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"#/components/schemas/Foo", "../schemas/Foo.json"},
		{"#/components/parameters/Bar", "../parameters/Bar.json"},
		{"#/components/responses/Baz", "../responses/Baz.json"},
		{"#/components/requestBodies/Qux", "../requestBodies/Qux.json"},
		{"#/components/headers/X-Rate-Limit", "../headers/X-Rate-Limit.json"},
		{"#/components/securitySchemes/BearerAuth", "../securitySchemes/BearerAuth.json"},
		{"#/components/links/UserLink", "../links/UserLink.json"},
		{"#/components/callbacks/cb", "../callbacks/cb.json"},
		{"#/components/pathItems/shared", "../pathItems/shared.json"},
		{"#/components/examples/PetExample", "../examples/PetExample.json"},
	}

	for _, tt := range tests {
		input := map[string]interface{}{"$ref": tt.ref}
		got := splitter.RewriteRefs(input, ".json", 1).(map[string]interface{})
		if got["$ref"] != tt.want {
			t.Errorf("RewriteRefs(%q) = %q, want %q", tt.ref, got["$ref"], tt.want)
		}
	}
}

func TestRewriteRefs_DepthZero(t *testing.T) {
	// At depth 0 (root file), refs use "./" prefix
	input := map[string]interface{}{
		"$ref": "#/components/schemas/Pet",
	}
	got := splitter.RewriteRefs(input, ".json", 0)
	m := got.(map[string]interface{})
	want := "./schemas/Pet.json"
	if m["$ref"] != want {
		t.Errorf("got $ref %q, want %q", m["$ref"], want)
	}
}

func TestRewriteRefs_UnknownRef(t *testing.T) {
	// Unknown refs are left unchanged
	input := map[string]interface{}{
		"$ref": "https://external.example.com/schema.json",
	}
	got := splitter.RewriteRefs(input, ".json", 1)
	m := got.(map[string]interface{})
	want := "https://external.example.com/schema.json"
	if m["$ref"] != want {
		t.Errorf("got $ref %q, want %q", m["$ref"], want)
	}
}

func TestRewriteRefs_Nested(t *testing.T) {
	// Verify refs are rewritten deep inside a nested structure
	input := map[string]interface{}{
		"get": map[string]interface{}{
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/Pet",
							},
						},
					},
				},
			},
		},
	}

	got := splitter.RewriteRefs(input, ".json", 1)
	// Navigate to the schema ref
	result := got.(map[string]interface{})["get"].(map[string]interface{})["responses"].(map[string]interface{})["200"].(map[string]interface{})["content"].(map[string]interface{})["application/json"].(map[string]interface{})["schema"].(map[string]interface{})["$ref"]
	want := "../schemas/Pet.json"
	if result != want {
		t.Errorf("nested ref = %q, want %q", result, want)
	}
}

func TestRewriteRefs_Array(t *testing.T) {
	// Verify refs inside array items are rewritten
	input := []interface{}{
		map[string]interface{}{"$ref": "#/components/schemas/Pet"},
		map[string]interface{}{"$ref": "#/components/schemas/Error"},
	}

	got := splitter.RewriteRefs(input, ".json", 1).([]interface{})
	if got[0].(map[string]interface{})["$ref"] != "../schemas/Pet.json" {
		t.Errorf("first array item: got %q", got[0].(map[string]interface{})["$ref"])
	}
	if got[1].(map[string]interface{})["$ref"] != "../schemas/Error.json" {
		t.Errorf("second array item: got %q", got[1].(map[string]interface{})["$ref"])
	}
}

func TestRewriteRefs_SiblingKeysPreserved(t *testing.T) {
	// In OpenAPI 3.1, $ref can have sibling keys like description
	input := map[string]interface{}{
		"$ref":        "#/components/schemas/Pet",
		"description": "Override description",
	}
	got := splitter.RewriteRefs(input, ".json", 1).(map[string]interface{})
	if got["$ref"] != "../schemas/Pet.json" {
		t.Errorf("$ref not rewritten: %q", got["$ref"])
	}
	if got["description"] != "Override description" {
		t.Errorf("sibling key lost: %q", got["description"])
	}
}
