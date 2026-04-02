// Package parser reads and validates OpenAPI specification files.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"gopkg.in/yaml.v3"
)

// ParsedDoc is the result of parsing an OpenAPI file.
type ParsedDoc struct {
	// Raw is the document as a generic map (used for splitting).
	Raw map[string]interface{}
	// Version is the detected OpenAPI version string (e.g. "3.0.3", "3.1.0").
	Version string
	// InputFormat is the detected input format: "json" or "yaml".
	InputFormat string
}

// Parse reads, validates and returns the OpenAPI document at the given path.
// It supports both JSON and YAML input and both OpenAPI 3.0.x and 3.1.x.
func Parse(filePath string, debug bool) (*ParsedDoc, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not resolve path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	if debug {
		fmt.Printf("[DEBUG] Parsing file: %s\n", absPath)
	}

	// Detect input format
	inputFormat := detectFormat(filePath, data)

	// Parse into generic map
	raw, err := parseRaw(data, inputFormat)
	if err != nil {
		return nil, fmt.Errorf("could not parse document: %w", err)
	}

	// Extract and validate the openapi version field
	version, err := extractVersion(raw)
	if err != nil {
		return nil, err
	}

	// Full schema validation via libopenapi
	if err := validate(data, debug); err != nil {
		return nil, fmt.Errorf("OpenAPI validation failed: %w", err)
	}

	return &ParsedDoc{
		Raw:         raw,
		Version:     version,
		InputFormat: inputFormat,
	}, nil
}

// detectFormat returns "json" or "yaml" based on file extension, falling back
// to content sniffing when the extension is ambiguous.
func detectFormat(filePath string, data []byte) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	default:
		// Sniff: JSON documents start with '{' (after optional whitespace)
		trimmed := strings.TrimSpace(string(data))
		if strings.HasPrefix(trimmed, "{") {
			return "json"
		}
		return "yaml"
	}
}

// parseRaw decodes the document into a map[string]interface{}.
func parseRaw(data []byte, format string) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if format == "json" {
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("JSON parse error: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("YAML parse error: %w", err)
		}
		// yaml.v3 may decode maps as map[string]interface{} or map[interface{}]interface{}.
		// Normalise to map[string]interface{} for consistency.
		normalised, err := normaliseMap(raw)
		if err != nil {
			return nil, err
		}
		raw = normalised
	}
	return raw, nil
}

// extractVersion pulls the "openapi" field and validates it is 3.0.x or 3.1.x.
func extractVersion(raw map[string]interface{}) (string, error) {
	v, ok := raw["openapi"]
	if !ok {
		return "", fmt.Errorf("missing required 'openapi' field")
	}
	vStr, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("'openapi' field must be a string")
	}
	if !strings.HasPrefix(vStr, "3.0") && !strings.HasPrefix(vStr, "3.1") {
		return "", fmt.Errorf("unsupported OpenAPI version %q: only 3.0.x and 3.1.x are supported", vStr)
	}
	return vStr, nil
}

// validate uses libopenapi to validate the document against the OpenAPI schema.
// It natively supports OpenAPI 3.0, 3.1, and 3.2 without any workarounds.
func validate(data []byte, debug bool) error {
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return fmt.Errorf("could not load document: %w", err)
	}

	docValidator, validatorErrs := validator.NewValidator(doc)
	if len(validatorErrs) > 0 {
		msgs := make([]string, 0, len(validatorErrs))
		for _, e := range validatorErrs {
			msgs = append(msgs, e.Error())
		}
		return fmt.Errorf("could not create validator: %s", strings.Join(msgs, "; "))
	}

	valid, validationErrs := docValidator.ValidateDocument()
	if !valid {
		msgs := make([]string, 0, len(validationErrs))
		for _, e := range validationErrs {
			msgs = append(msgs, e.Message)
		}
		if debug {
			for _, e := range validationErrs {
				fmt.Printf("[DEBUG] Validation error: %s — %s\n", e.ValidationType, e.Message)
			}
		}
		return fmt.Errorf("document validation failed: %s", strings.Join(msgs, "; "))
	}

	return nil
}

// normaliseMap recursively converts map[interface{}]interface{} (produced by
// some YAML decoders) to map[string]interface{}.
func normaliseMap(in map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		normalised, err := normaliseValue(v)
		if err != nil {
			return nil, err
		}
		out[k] = normalised
	}
	return out, nil
}

func normaliseValue(v interface{}) (interface{}, error) {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(val))
		for mk, mv := range val {
			key, ok := mk.(string)
			if !ok {
				return nil, fmt.Errorf("non-string map key: %v", mk)
			}
			normalised, err := normaliseValue(mv)
			if err != nil {
				return nil, err
			}
			m[key] = normalised
		}
		return m, nil
	case map[string]interface{}:
		return normaliseMap(val)
	case []interface{}:
		for i, item := range val {
			normalised, err := normaliseValue(item)
			if err != nil {
				return nil, err
			}
			val[i] = normalised
		}
		return val, nil
	default:
		return v, nil
	}
}
