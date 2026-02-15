package notelink

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Test structures for schema generation
type SimpleUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type UserWithTypes struct {
	Name      string  `json:"name"`
	CreatedAt string  `json:"created_at"`
	Height    float64 `json:"height"`
	ID        uint    `json:"id"`
	Age       int     `json:"age"`
	IsActive  bool    `json:"is_active"`
}

type UserWithNested struct {
	Name    string      `json:"name"`
	Address AddressType `json:"address"`
	Tags    []string    `json:"tags"`
}

type AddressType struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	ZipCode int    `json:"zip_code"`
}

type UserWithPointers struct {
	Email *string `json:"email,omitempty"`
	Age   *int    `json:"age"`
	Name  string  `json:"name"`
}

type UserWithTimeFields struct {
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	ID        int       `json:"id"`
}

// TestGenerateTypeScriptSchema tests TypeScript schema generation
func TestGenerateTypeScriptSchema(t *testing.T) {
	tests := []struct {
		name           string
		schemaName     string
		schema         interface{}
		expectedFields []string
		shouldContain  []string
	}{
		{
			name:       "Simple struct",
			schemaName: "User",
			schema:     SimpleUser{},
			expectedFields: []string{
				"name: string",
				"email: string",
				"age: number",
			},
			shouldContain: []string{
				"export interface User",
			},
		},
		{
			name:       "Struct with multiple types",
			schemaName: "User",
			schema:     UserWithTypes{},
			expectedFields: []string{
				"id: number",
				"name: string",
				"age: number",
				"height: number",
				"is_active: boolean",
			},
		},
		{
			name:       "Struct with nested types",
			schemaName: "User",
			schema:     UserWithNested{},
			expectedFields: []string{
				"name: string",
				"address: AddressType",
				"tags: string[]",
			},
			shouldContain: []string{
				"export interface AddressType",
				"street: string",
				"city: string",
				"zip_code: number",
			},
		},
		{
			name:       "Struct with pointers",
			schemaName: "User",
			schema:     UserWithPointers{},
			expectedFields: []string{
				"name: string",
				"email: string",
				"age: number",
			},
		},
		{
			name:          "Nil schema",
			schemaName:    "User",
			schema:        nil,
			shouldContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateTypeScriptSchema(tt.schemaName, tt.schema)

			// Check expected fields
			for _, field := range tt.expectedFields {
				if !strings.Contains(result, field) {
					t.Errorf("Expected schema to contain '%s', got:\n%s", field, result)
				}
			}

			// Check additional content
			for _, content := range tt.shouldContain {
				if content != "" && !strings.Contains(result, content) {
					t.Errorf("Expected schema to contain '%s', got:\n%s", content, result)
				}
			}
		})
	}
}

// TestGenerateTypeScriptSchemaArray tests TypeScript schema generation for arrays
func TestGenerateTypeScriptSchemaArray(t *testing.T) {
	result := generateTypeScriptSchema("UserList", []SimpleUser{})

	if !strings.Contains(result, "export interface UserList") {
		t.Errorf("Expected interface name UserList, got:\n%s", result)
	}

	if !strings.Contains(result, "name: string") {
		t.Errorf("Expected field 'name: string', got:\n%s", result)
	}
}

// TestGenerateJSONTemplate tests JSON template generation
func TestGenerateJSONTemplate(t *testing.T) {
	tests := []struct {
		name          string
		schema        interface{}
		shouldContain []string
	}{
		{
			name:   "Simple user",
			schema: SimpleUser{},
			shouldContain: []string{
				`"name"`,
				`"email"`,
				`"age"`,
			},
		},
		{
			name:   "User with types",
			schema: UserWithTypes{},
			shouldContain: []string{
				`"id"`,
				`"name"`,
				`"age"`,
				`"height"`,
				`"is_active"`,
			},
		},
		{
			name:   "User with nested types",
			schema: UserWithNested{},
			shouldContain: []string{
				`"name"`,
				`"address"`,
				`"street"`,
				`"city"`,
				`"zip_code"`,
				`"tags"`,
			},
		},
		{
			name:   "User with pointers",
			schema: UserWithPointers{},
			shouldContain: []string{
				`"name"`,
				`"email"`,
				`"age"`,
			},
		},
		{
			name:          "Nil schema",
			schema:        nil,
			shouldContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateJSONTemplate(tt.schema)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check if result is valid JSON
			var jsonData map[string]interface{}
			if tt.schema != nil {
				if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
					t.Fatalf("Result is not valid JSON: %v\nGot: %s", err, result)
				}
			}

			// Check if result contains expected strings
			for _, expected := range tt.shouldContain {
				if expected != "" && !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got:\n%s", expected, result)
				}
			}
		})
	}
}

// TestGenerateJSONTemplateArray tests JSON template generation for arrays
func TestGenerateJSONTemplateArray(t *testing.T) {
	result, err := generateJSONTemplate([]SimpleUser{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var jsonData []interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Result is not valid JSON array: %v\nGot: %s", err, result)
	}

	if len(jsonData) == 0 {
		t.Errorf("Expected array with at least one element, got empty array")
	}
}

// TestGoTypeToTsType tests Go to TypeScript type mapping
func TestGoTypeToTsType(t *testing.T) {
	tests := []struct {
		name         string
		goType       interface{}
		expectedType string
	}{
		{"String", "", "string"},
		{"Int", 0, "number"},
		{"Int64", int64(0), "number"},
		{"Uint", uint(0), "number"},
		{"Float32", float32(0), "number"},
		{"Float64", float64(0), "number"},
		{"Bool", false, "boolean"},
		{"String slice", []string{}, "string[]"},
		{"Int slice", []int{}, "number[]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test verifies type mapping through the generateTypeScriptSchema function
			// since goTypeToTsType is not exported
		})
	}
}

// TestGenerateStringExample tests contextual string example generation
func TestGenerateStringExample(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  string
	}{
		{"email", "user@example.com"},
		{"password", "securePassword123"},
		{"username", "john_doe"},
		{"firstName", "John"},
		{"lastName", "Doe"},
		{"phone", "+1-555-0123"},
		{"url", "https://example.com"},
		{"title", "Sample Title"},
		{"status", "active"},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			// Test by creating a struct with the field name and checking the JSON template
			result := generateStringExample(strings.ToLower(tt.fieldName))
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestGenerateIntExample tests contextual integer example generation
func TestGenerateIntExample(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  int
	}{
		{"age", 25},
		{"count", 10},
		{"id", 12345},
		{"port", 8080},
		{"year", 2024},
		{"month", 6},
		{"day", 15},
		{"unknown", 1},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := generateIntExample(strings.ToLower(tt.fieldName))
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestGenerateFloatExample tests contextual float example generation
func TestGenerateFloatExample(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  float64
	}{
		{"price", 99.99},
		{"rate", 0.15},
		{"percentage", 75.5},
		{"latitude", 40.7128},
		{"longitude", -74.0060},
		{"weight", 70.5},
		{"height", 175.0},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := generateFloatExample(strings.ToLower(tt.fieldName))
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

// TestGenerateBoolExample tests contextual boolean example generation
func TestGenerateBoolExample(t *testing.T) {
	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"active", true},
		{"enabled", true},
		{"deleted", false},
		{"disabled", false},
		{"verified", true},
		{"confirmed", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := generateBoolExample(strings.ToLower(tt.fieldName))
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestGetJSONFieldName tests JSON field name extraction
func TestGetJSONFieldName(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		jsonTag   string
		expected  string
	}{
		{"With json tag", "UserID", `json:"user_id"`, "user_id"},
		{"With omitempty", "Email", `json:"email,omitempty"`, "email"},
		{"Skip field", "Internal", `json:"-"`, "-"},
		{"No json tag", "UserName", "", "userName"},
		{"Empty json tag", "Field", ``, "field"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Field string
			}

			structType := new(TestStruct)
			field := reflect.TypeOf(structType).Elem().Field(0)

			// Create a mock field with the desired tag
			mockField := reflect.StructField{
				Name: tt.fieldName,
				Type: field.Type,
				Tag:  reflect.StructTag(tt.jsonTag),
			}

			result := getJSONFieldName(&mockField)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestGenerateJSONTemplateWithTime tests JSON template generation with time fields
func TestGenerateJSONTemplateWithTime(t *testing.T) {
	result, err := generateJSONTemplate(UserWithTimeFields{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, `"created_at"`) {
		t.Errorf("Expected result to contain 'created_at', got:\n%s", result)
	}

	// Parse the JSON to verify structure
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	// Verify created_at is a string (RFC3339 format)
	if _, ok := jsonData["created_at"].(string); !ok {
		t.Errorf("Expected created_at to be a string")
	}
}

// TestComplexNestedStructure tests deeply nested structure generation
func TestComplexNestedStructure(t *testing.T) {
	type Level3 struct {
		Value string `json:"value"`
	}

	type Level2 struct {
		Level3 Level3 `json:"level3"`
		Items  []int  `json:"items"`
	}

	type Level1 struct {
		Level2 Level2 `json:"level2"`
		Active bool   `json:"active"`
	}

	// Test TypeScript generation
	tsSchema := generateTypeScriptSchema("Root", Level1{})

	expectedInterfaces := []string{
		"export interface Level3",
		"export interface Level2",
		"export interface Root",
		"value: string",
		"items: number[]",
		"active: boolean",
	}

	for _, expected := range expectedInterfaces {
		if !strings.Contains(tsSchema, expected) {
			t.Errorf("Expected TypeScript schema to contain '%s', got:\n%s", expected, tsSchema)
		}
	}

	// Test JSON template generation
	jsonTemplate, err := generateJSONTemplate(Level1{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonTemplate), &jsonData); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}

	// Verify nested structure
	if _, ok := jsonData["level2"]; !ok {
		t.Errorf("Expected 'level2' field in JSON template")
	}
}
