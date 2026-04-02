// Package orderedmap provides an ordered map that preserves insertion order
// when serialised to JSON or YAML.
package orderedmap

import (
	"bytes"
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// OrderedMap is a map that preserves insertion order during JSON and YAML serialisation.
type OrderedMap struct {
	keys   []string
	values []interface{}
	index  map[string]int // key → position in keys/values slices
}

// NewOrderedMap returns an empty OrderedMap.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		index: make(map[string]int),
	}
}

// Set inserts or updates a key/value pair. If the key already exists its
// value is updated in place without changing its position. New keys are
// appended at the end.
func (m *OrderedMap) Set(key string, value interface{}) {
	if idx, ok := m.index[key]; ok {
		m.values[idx] = value
		return
	}
	m.index[key] = len(m.keys)
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
}

// Get returns the value for the given key and whether the key was present.
func (m *OrderedMap) Get(key string) (interface{}, bool) {
	idx, ok := m.index[key]
	if !ok {
		return nil, false
	}
	return m.values[idx], true
}

// Keys returns the keys in insertion order.
func (m *OrderedMap) Keys() []string {
	return m.keys
}

// Len returns the number of entries.
func (m *OrderedMap) Len() int {
	return len(m.keys)
}

// MarshalJSON implements json.Marshaler. Keys are emitted in insertion order.
func (m *OrderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')
		valBytes, err := json.Marshal(m.values[i])
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// MarshalYAML implements yaml.Marshaler. Keys are emitted in insertion order.
func (m *OrderedMap) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
	for i, key := range m.keys {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		}
		valueNode, err := toYAMLNode(m.values[i])
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, keyNode, valueNode)
		_ = i
	}
	return node, nil
}

// toYAMLNode converts a Go value to a *yaml.Node via marshal/unmarshal round-trip.
func toYAMLNode(v interface{}) (*yaml.Node, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
}
