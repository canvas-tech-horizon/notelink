package notelink

import (
	"reflect"
	"strings"
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
