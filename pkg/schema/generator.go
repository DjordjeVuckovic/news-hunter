package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	_ "regexp"
	"strconv"
	"strings"
)

// JSONSchema represents a JSON Schema document
type JSONSchema struct {
	Schema      string                 `json:"$schema"`
	ID          string                 `json:"$id,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"`
	Required    []string               `json:"required,omitempty"`
	Properties  map[string]*JSONSchema `json:"properties,omitempty"`
	Items       *JSONSchema            `json:"items,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	MinLength   *int                   `json:"minLength,omitempty"`
	MaxLength   *int                   `json:"maxLength,omitempty"`
	MinItems    *int                   `json:"minItems,omitempty"`
	MaxItems    *int                   `json:"maxItems,omitempty"`
	Examples    []interface{}          `json:"examples,omitempty"`
}

const schemaRef = "https://json-schema.org/draft/2020-12/schema"

// Generator generates JSON schemas from Go structs
type Generator struct {
	schemas map[string]*JSONSchema
}

// NewGenerator creates a new schema generator
func NewGenerator() *Generator {
	return &Generator{
		schemas: make(map[string]*JSONSchema),
	}
}

// GenerateSchema generates a JSON schema from a Go type
func (g *Generator) GenerateSchema(t reflect.Type) (*JSONSchema, error) {
	return g.generateSchemaForType(t, true)
}

func (g *Generator) generateSchemaForType(t reflect.Type, isRoot bool) (*JSONSchema, error) {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := &JSONSchema{
		Schema: schemaRef,
	}

	switch t.Kind() {
	case reflect.Struct:
		return g.generateStructSchema(t, isRoot)
	case reflect.Slice:
		return g.generateSliceSchema(t)
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = "integer"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	default:
		return nil, fmt.Errorf("unsupported type: %schemaRef", t.Kind())
	}

	return schema, nil
}

func (g *Generator) generateStructSchema(t reflect.Type, isRoot bool) (*JSONSchema, error) {
	schema := &JSONSchema{
		Type:       "object",
		Properties: make(map[string]*JSONSchema),
	}

	if isRoot {
		schema.Schema = schemaRef

		// Extract schema metadata from type comments
		if comment := g.getTypeComment(t); comment != "" {
			schema.Description = comment
		}

		// Parse schema annotations from type comments
		g.parseSchemaAnnotations(t, schema)
	}

	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := g.getFieldName(field)
		if fieldName == "" {
			continue
		}

		fieldSchema, err := g.generateFieldSchema(field)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema for field %schemaRef: %w", field.Name, err)
		}

		schema.Properties[fieldName] = fieldSchema

		// Check if field is required
		if g.isFieldRequired(field) {
			required = append(required, fieldName)
		}
	}

	if len(required) > 0 {
		schema.Required = required
	}

	return schema, nil
}

func (g *Generator) generateSliceSchema(t reflect.Type) (*JSONSchema, error) {
	schema := &JSONSchema{
		Type: "array",
	}

	elemType := t.Elem()
	itemSchema, err := g.generateSchemaForType(elemType, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schema for array items: %w", err)
	}

	schema.Items = itemSchema
	return schema, nil
}

func (g *Generator) generateFieldSchema(field reflect.StructField) (*JSONSchema, error) {
	fieldSchema, err := g.generateSchemaForType(field.Type, false)
	if err != nil {
		return nil, err
	}

	// Add description from field comment
	if desc := field.Tag.Get("description"); desc != "" {
		fieldSchema.Description = desc
	}

	// Parse schema tag
	if schemaTag := field.Tag.Get("schema"); schemaTag != "" {
		g.parseSchemaTag(schemaTag, fieldSchema)
	}

	return fieldSchema, nil
}

func (g *Generator) parseSchemaTag(tag string, schema *JSONSchema) {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part == "required" {
			// Required is handled at the struct level
			continue
		}

		if strings.HasPrefix(part, "enum=") {
			enumStr := strings.TrimPrefix(part, "enum=")
			enums := strings.Split(enumStr, "|")
			schema.Enum = make([]interface{}, len(enums))
			for i, e := range enums {
				schema.Enum[i] = e
			}
		}

		if strings.HasPrefix(part, "default=") {
			defaultStr := strings.TrimPrefix(part, "default=")
			schema.Default = defaultStr
		}

		if strings.HasPrefix(part, "pattern=") {
			schema.Pattern = strings.TrimPrefix(part, "pattern=")
		}

		if strings.HasPrefix(part, "minLength=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(part, "minLength=")); err == nil {
				schema.MinLength = &val
			}
		}

		if strings.HasPrefix(part, "maxLength=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(part, "maxLength=")); err == nil {
				schema.MaxLength = &val
			}
		}

		if strings.HasPrefix(part, "minItems=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(part, "minItems=")); err == nil {
				schema.MinItems = &val
			}
		}

		if strings.HasPrefix(part, "maxItems=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(part, "maxItems=")); err == nil {
				schema.MaxItems = &val
			}
		}
	}
}

func (g *Generator) getFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return strings.ToLower(field.Name[:1]) + field.Name[1:]
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] == "" {
		return strings.ToLower(field.Name[:1]) + field.Name[1:]
	}

	return parts[0]
}

func (g *Generator) isFieldRequired(field reflect.StructField) bool {
	schemaTag := field.Tag.Get("schema")
	return strings.Contains(schemaTag, "required")
}

func (g *Generator) getTypeComment(t reflect.Type) string {
	// This would typically come from parsing the source file
	// For now, return empty string
	return ""
}

func (g *Generator) parseSchemaAnnotations(t reflect.Type, schema *JSONSchema) {
	// Parse annotations like +schema:root=true, +schema:group=newshunter.io
	// In a real implementation, this would parse the source file comments
	// For now, we'll set some defaults
	schema.Title = t.Name()
	schema.ID = fmt.Sprintf("https://schemas.newshunter.io/%s", strings.ToLower(t.Name()))
}

// GenerateJSONSchema generates a JSON schema as a JSON string
func (g *Generator) GenerateJSONSchema(v interface{}) (string, error) {
	t := reflect.TypeOf(v)
	schema, err := g.GenerateSchema(t)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema to JSON: %w", err)
	}

	return string(jsonBytes), nil
}
