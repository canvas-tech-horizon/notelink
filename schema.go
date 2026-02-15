package notelink

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

// generateTypeScriptSchema converts a Go type to TypeScript interfaces, including nested structs
func generateTypeScriptSchema(name string, schema interface{}) string {
	if schema == nil {
		return ""
	}

	typ := reflect.TypeOf(schema)
	if typ == nil {
		return ""
	}

	var ts strings.Builder
	seenTypes := make(map[string]bool) // To avoid duplicate definitions
	isArray := false

	// Handle pointers and slices
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		isArray = true
	}

	if typ.Kind() != reflect.Struct {
		return ""
	}

	// Generate all nested structs first
	generateAllStructs(typ, &ts, seenTypes)

	// Generate the main interface
	ts.WriteString(`export interface ` + name + " {\n")
	ts.WriteString(generateStructSchema(typ))
	ts.WriteString("}")

	if isArray {
		return ts.String() + "[]"
	}
	return ts.String()
}

// generateAllStructs recursively generates interfaces for all nested structs
func generateAllStructs(typ reflect.Type, ts *strings.Builder, seenTypes map[string]bool) {
	if typ.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type

		// Handle pointer and slice types
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct && !seenTypes[fieldType.Name()] && fieldType.Name() != "" {
			seenTypes[fieldType.Name()] = true
			// Recursively generate nested structs
			generateAllStructs(fieldType, ts, seenTypes)
			// Generate the interface for this struct
			ts.WriteString(`export interface ` + fieldType.Name() + " {\n")
			ts.WriteString(generateStructSchema(fieldType))
			ts.WriteString("}\n\n")
		}
	}
}

// generateStructSchema generates TypeScript for a struct type
func generateStructSchema(typ reflect.Type) string {
	var ts strings.Builder
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		tsType := goTypeToTsType(fieldType)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			fieldName = parts[0]
		} else {
			// Default to camelCase if no JSON tag
			fieldName = strings.ToLower(fieldName[:1]) + fieldName[1:]
		}
		ts.WriteString("  " + fieldName + ": " + tsType + ";\n")
	}
	return ts.String()
}

// goTypeToTsType maps Go types to TypeScript types
func goTypeToTsType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice:
		return goTypeToTsType(t.Elem()) + "[]"
	case reflect.Ptr:
		return goTypeToTsType(t.Elem())
	case reflect.Struct:
		if t.Name() == "" {
			return "any" // Anonymous structs
		}
		return t.Name() // Named structs
	default:
		return "any"
	}
}

// generateJSONTemplate creates a JSON template from a Go struct schema
func generateJSONTemplate(schema interface{}) (string, error) {
	if schema == nil {
		return "{}", nil
	}

	template := generateJSONFromType(reflect.TypeOf(schema))
	jsonBytes, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return "{}", err
	}

	return string(jsonBytes), nil
}

// generateJSONFromType recursively creates example JSON data from a reflect.Type
func generateJSONFromType(t reflect.Type) interface{} {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		return generateJSONFromType(t.Elem())
	}

	// Handle slices
	if t.Kind() == reflect.Slice {
		elemExample := generateJSONFromType(t.Elem())
		return []interface{}{elemExample}
	}

	// Handle structs
	if t.Kind() == reflect.Struct {
		result := make(map[string]interface{})

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Get field name from JSON tag or use field name
			fieldName := getJSONFieldName(&field)
			if fieldName == "-" {
				continue // Skip fields marked with json:"-"
			}

			// Generate example value for this field
			result[fieldName] = generateExampleValue(field.Type, field.Name)
		}

		return result
	}

	// For non-struct types, generate example values
	return generateExampleValue(t, "")
}

// getJSONFieldName extracts the JSON field name from struct field tags
func getJSONFieldName(field *reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		// Convert to camelCase if no JSON tag
		name := field.Name
		return strings.ToLower(name[:1]) + name[1:]
	}

	// Parse JSON tag (e.g., "field_name,omitempty")
	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

// generateExampleValue creates example values based on type and field name
func generateExampleValue(t reflect.Type, fieldName string) interface{} {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		return generateExampleValue(t.Elem(), fieldName)
	}

	// Handle slices
	if t.Kind() == reflect.Slice {
		elemExample := generateExampleValue(t.Elem(), fieldName)
		return []interface{}{elemExample}
	}

	// Handle structs
	if t.Kind() == reflect.Struct {
		// Special case for time.Time
		if t == reflect.TypeOf(time.Time{}) {
			return time.Now().Format(time.RFC3339)
		}
		return generateJSONFromType(t)
	}

	// Generate examples based on field name patterns and types
	lowerFieldName := strings.ToLower(fieldName)

	switch t.Kind() {
	case reflect.String:
		return generateStringExample(lowerFieldName)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return generateIntExample(lowerFieldName)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return generateUintExample(lowerFieldName)
	case reflect.Float32, reflect.Float64:
		return generateFloatExample(lowerFieldName)
	case reflect.Bool:
		return generateBoolExample(lowerFieldName)
	default:
		return nil
	}
}

// generateStringExample creates contextual string examples based on field names
func generateStringExample(fieldName string) string {
	switch {
	case strings.Contains(fieldName, "email"):
		return "user@example.com"
	case strings.Contains(fieldName, "password"):
		return "securePassword123"
	case strings.Contains(fieldName, "username") || strings.Contains(fieldName, "user_name"):
		return "john_doe"
	case strings.Contains(fieldName, "firstname") || strings.Contains(fieldName, "first_name"):
		return "John"
	case strings.Contains(fieldName, "lastname") || strings.Contains(fieldName, "last_name"):
		return "Doe"
	case strings.Contains(fieldName, "name"):
		return "John Doe"
	case strings.Contains(fieldName, "phone"):
		return "+1-555-0123"
	case strings.Contains(fieldName, "address"):
		return "123 Main Street, City, Country"
	case strings.Contains(fieldName, "url") || strings.Contains(fieldName, "link"):
		return "https://example.com"
	case strings.Contains(fieldName, "id"):
		return "12345"
	case strings.Contains(fieldName, "token"):
		return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
	case strings.Contains(fieldName, "description"):
		return "This is a sample description"
	case strings.Contains(fieldName, "title"):
		return "Sample Title"
	case strings.Contains(fieldName, "status"):
		return "active"
	case strings.Contains(fieldName, "type"):
		return "default"
	default:
		return "example_value"
	}
}

// generateIntExample creates contextual integer examples
func generateIntExample(fieldName string) int {
	switch {
	case strings.Contains(fieldName, "age"):
		return 25
	case strings.Contains(fieldName, "count") || strings.Contains(fieldName, "total"):
		return 10
	case strings.Contains(fieldName, "id"):
		return 12345
	case strings.Contains(fieldName, "port"):
		return 8080
	case strings.Contains(fieldName, "year"):
		return 2024
	case strings.Contains(fieldName, "month"):
		return 6
	case strings.Contains(fieldName, "day"):
		return 15
	default:
		return 1
	}
}

// generateUintExample creates contextual unsigned integer examples
func generateUintExample(fieldName string) uint {
	intVal := generateIntExample(fieldName)
	if intVal < 0 {
		return 0
	}
	return uint(intVal)
}

// generateFloatExample creates contextual float examples
func generateFloatExample(fieldName string) float64 {
	switch {
	case strings.Contains(fieldName, "price") || strings.Contains(fieldName, "cost"):
		return 99.99
	case strings.Contains(fieldName, "rate"):
		return 0.15
	case strings.Contains(fieldName, "percentage"):
		return 75.5
	case strings.Contains(fieldName, "latitude"):
		return 40.7128
	case strings.Contains(fieldName, "longitude"):
		return -74.0060
	case strings.Contains(fieldName, "weight"):
		return 70.5
	case strings.Contains(fieldName, "height"):
		return 175.0
	default:
		return 1.0
	}
}

// generateBoolExample creates contextual boolean examples
func generateBoolExample(fieldName string) bool {
	switch {
	case strings.Contains(fieldName, "active") || strings.Contains(fieldName, "enabled"):
		return true
	case strings.Contains(fieldName, "deleted") || strings.Contains(fieldName, "disabled"):
		return false
	case strings.Contains(fieldName, "verified") || strings.Contains(fieldName, "confirmed"):
		return true
	default:
		return false
	}
}
