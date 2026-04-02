package orderedmap_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/akyrey/openapi-splitter/internal/orderedmap"
	"gopkg.in/yaml.v3"
)

func TestOrderedMap_JSON_KeyOrder(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	m.Set("z", "last")
	m.Set("a", "first")
	m.Set("m", "middle")

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	keys := decodeJSONKeys(t, dec)
	assertKeyOrder(t, keys, []string{"z", "a", "m"})
}

func TestOrderedMap_JSON_SetUpdatesInPlace(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("a", 99) // update — should not change order

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	keys := decodeJSONKeys(t, dec)
	assertKeyOrder(t, keys, []string{"a", "b"})

	v, ok := m.Get("a")
	if !ok || v != 99 {
		t.Errorf("expected a=99, got %v (ok=%v)", v, ok)
	}
}

func TestOrderedMap_JSON_Empty(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(data) != "{}" {
		t.Errorf("expected {}, got %s", data)
	}
}

func TestOrderedMap_JSON_NestedValues(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	m.Set("nested", map[string]interface{}{"x": 1})
	m.Set("arr", []interface{}{1, 2, 3})

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
}

func TestOrderedMap_YAML_KeyOrder(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	m.Set("openapi", "3.1.0")
	m.Set("info", map[string]interface{}{"title": "Test"})
	m.Set("paths", map[string]interface{}{})

	data, err := yaml.Marshal(m)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}

	keys := decodeYAMLTopLevelKeys(t, data)
	assertKeyOrder(t, keys, []string{"openapi", "info", "paths"})
}

func TestOrderedMap_YAML_Empty(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	data, err := yaml.Marshal(m)
	if err != nil {
		t.Fatalf("yaml.Marshal: %v", err)
	}
	s := string(data)
	if s != "{}\n" {
		t.Errorf("expected {}\\n, got %q", s)
	}
}

func TestOrderedMap_Get_Missing(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	v, ok := m.Get("missing")
	if ok || v != nil {
		t.Errorf("expected (nil, false), got (%v, %v)", v, ok)
	}
}

func TestOrderedMap_Len(t *testing.T) {
	m := orderedmap.NewOrderedMap()
	if m.Len() != 0 {
		t.Errorf("expected 0, got %d", m.Len())
	}
	m.Set("a", 1)
	m.Set("b", 2)
	if m.Len() != 2 {
		t.Errorf("expected 2, got %d", m.Len())
	}
	m.Set("a", 99) // update should not increase Len
	if m.Len() != 2 {
		t.Errorf("expected 2 after update, got %d", m.Len())
	}
}

// --- helpers ---

func decodeJSONKeys(t *testing.T, dec *json.Decoder) []string {
	t.Helper()
	tok, err := dec.Token()
	if err != nil {
		t.Fatalf("expected '{': %v", err)
	}
	if tok != json.Delim('{') {
		t.Fatalf("expected '{', got %v", tok)
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

func decodeYAMLTopLevelKeys(t *testing.T, data []byte) []string {
	t.Helper()
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		t.Fatalf("yaml.Unmarshal: %v", err)
	}
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		t.Fatalf("expected DocumentNode with content")
	}
	mapping := node.Content[0]
	if mapping.Kind != yaml.MappingNode {
		t.Fatalf("expected MappingNode, got %v", mapping.Kind)
	}
	var keys []string
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		keys = append(keys, mapping.Content[i].Value)
	}
	return keys
}

func assertKeyOrder(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("key count: got %d (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("key[%d]: got %q, want %q", i, got[i], want[i])
		}
	}
}
