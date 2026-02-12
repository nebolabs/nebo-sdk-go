package nebo

import "encoding/json"

// SchemaBuilder constructs JSON Schema for STRAP-pattern tool inputs.
type SchemaBuilder struct {
	actions    []string
	properties map[string]map[string]interface{}
	required   []string
}

// NewSchema creates a SchemaBuilder with the given action names.
// Actions become an enum on the required "action" field.
func NewSchema(actions ...string) *SchemaBuilder {
	return &SchemaBuilder{
		actions:    actions,
		properties: make(map[string]map[string]interface{}),
	}
}

// String adds a string parameter to the schema.
func (s *SchemaBuilder) String(name, description string, required bool) *SchemaBuilder {
	s.properties[name] = map[string]interface{}{
		"type":        "string",
		"description": description,
	}
	if required {
		s.required = append(s.required, name)
	}
	return s
}

// Number adds a number parameter to the schema.
func (s *SchemaBuilder) Number(name, description string, required bool) *SchemaBuilder {
	s.properties[name] = map[string]interface{}{
		"type":        "number",
		"description": description,
	}
	if required {
		s.required = append(s.required, name)
	}
	return s
}

// Bool adds a boolean parameter to the schema.
func (s *SchemaBuilder) Bool(name, description string, required bool) *SchemaBuilder {
	s.properties[name] = map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
	if required {
		s.required = append(s.required, name)
	}
	return s
}

// Enum adds a string enum parameter to the schema.
func (s *SchemaBuilder) Enum(name, description string, required bool, values ...string) *SchemaBuilder {
	s.properties[name] = map[string]interface{}{
		"type":        "string",
		"enum":        values,
		"description": description,
	}
	if required {
		s.required = append(s.required, name)
	}
	return s
}

// Object adds an object parameter to the schema.
func (s *SchemaBuilder) Object(name, description string, required bool) *SchemaBuilder {
	s.properties[name] = map[string]interface{}{
		"type":        "object",
		"description": description,
	}
	if required {
		s.required = append(s.required, name)
	}
	return s
}

// Build returns the complete JSON Schema as json.RawMessage.
func (s *SchemaBuilder) Build() json.RawMessage {
	props := make(map[string]interface{})

	// Action is always the first property and always required
	props["action"] = map[string]interface{}{
		"type":        "string",
		"enum":        s.actions,
		"description": "Action to perform: " + joinStrings(s.actions),
	}

	for k, v := range s.properties {
		props[k] = v
	}

	required := append([]string{"action"}, s.required...)

	schema := map[string]interface{}{
		"type":       "object",
		"properties": props,
		"required":   required,
	}

	data, _ := json.Marshal(schema)
	return data
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += ", " + s
	}
	return result
}
