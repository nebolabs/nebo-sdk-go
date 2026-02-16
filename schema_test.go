package nebo

import (
	"encoding/json"
	"testing"
)

func TestSchemaBuilderBasic(t *testing.T) {
	schema := NewSchema("add", "subtract").
		Number("a", "First operand", true).
		Number("b", "Second operand", true).
		Build()

	var parsed map[string]any
	if err := json.Unmarshal(schema, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if parsed["type"] != "object" {
		t.Errorf("type = %v, want object", parsed["type"])
	}

	props := parsed["properties"].(map[string]any)
	if _, ok := props["action"]; !ok {
		t.Error("missing 'action' property")
	}
	if _, ok := props["a"]; !ok {
		t.Error("missing 'a' property")
	}
	if _, ok := props["b"]; !ok {
		t.Error("missing 'b' property")
	}

	// Check action enum
	actionProp := props["action"].(map[string]any)
	actionEnum := actionProp["enum"].([]any)
	if len(actionEnum) != 2 {
		t.Errorf("action enum length = %d, want 2", len(actionEnum))
	}
	if actionEnum[0] != "add" || actionEnum[1] != "subtract" {
		t.Errorf("action enum = %v, want [add, subtract]", actionEnum)
	}

	// Check required includes action and both operands
	required := parsed["required"].([]any)
	requiredSet := make(map[string]bool)
	for _, r := range required {
		requiredSet[r.(string)] = true
	}
	for _, field := range []string{"action", "a", "b"} {
		if !requiredSet[field] {
			t.Errorf("expected %q in required", field)
		}
	}
}

func TestSchemaBuilderAllTypes(t *testing.T) {
	schema := NewSchema("run").
		String("name", "Resource name", true).
		Number("count", "How many", false).
		Bool("verbose", "Show details", false).
		Enum("format", "Output format", true, "json", "text").
		Object("config", "Configuration", false).
		Build()

	var parsed map[string]any
	if err := json.Unmarshal(schema, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	props := parsed["properties"].(map[string]any)

	tests := []struct {
		name     string
		propType string
	}{
		{"name", "string"},
		{"count", "number"},
		{"verbose", "boolean"},
		{"format", "string"},
		{"config", "object"},
	}

	for _, tt := range tests {
		prop, ok := props[tt.name].(map[string]any)
		if !ok {
			t.Errorf("missing property %q", tt.name)
			continue
		}
		if prop["type"] != tt.propType {
			t.Errorf("%s.type = %v, want %s", tt.name, prop["type"], tt.propType)
		}
	}

	// Check format has enum values
	formatProp := props["format"].(map[string]any)
	formatEnum := formatProp["enum"].([]any)
	if len(formatEnum) != 2 {
		t.Errorf("format enum length = %d, want 2", len(formatEnum))
	}

	// Check required: action, name, format (not count, verbose, config)
	required := parsed["required"].([]any)
	requiredSet := make(map[string]bool)
	for _, r := range required {
		requiredSet[r.(string)] = true
	}
	if !requiredSet["action"] || !requiredSet["name"] || !requiredSet["format"] {
		t.Error("expected action, name, format in required")
	}
	if requiredSet["count"] || requiredSet["verbose"] || requiredSet["config"] {
		t.Error("count, verbose, config should not be required")
	}
}

func TestSchemaBuilderEmpty(t *testing.T) {
	schema := NewSchema("list").Build()

	var parsed map[string]any
	if err := json.Unmarshal(schema, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	props := parsed["properties"].(map[string]any)
	if len(props) != 1 {
		t.Errorf("expected only 'action' property, got %d properties", len(props))
	}
}
