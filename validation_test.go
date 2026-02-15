package notelink

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// Test structures for validation
type TestUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type TestUserWithOptional struct {
	Age      *int   `json:"age"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	IsActive bool   `json:"is_active"`
}

type TestNestedStruct struct {
	User    TestUser `json:"user"`
	Tags    []string `json:"tags"`
	Enabled bool     `json:"enabled"`
}

type TestArrayOfStructs struct {
	Users []TestUser `json:"users"`
}

type TestDeepNested struct {
	Level1 struct {
		Level2 struct {
			Level3 struct {
				Value string `json:"value"`
			} `json:"level3"`
		} `json:"level2"`
	} `json:"level1"`
}

// TestValidateParameters tests parameter validation for different types and locations
func TestValidateParameters(t *testing.T) {
	tests := []struct {
		headers     map[string]string
		url         string
		name        string
		params      []Parameter
		errorCount  int
		expectError bool
	}{
		{
			name: "Valid string parameter",
			params: []Parameter{
				{Name: "name", In: "query", Type: "string", Required: true},
			},
			url:         "/test?name=John",
			expectError: false,
		},
		{
			name: "Missing required parameter",
			params: []Parameter{
				{Name: "name", In: "query", Type: "string", Required: true},
			},
			url:         "/test",
			expectError: true,
			errorCount:  1,
		},
		{
			name: "Valid integer parameter",
			params: []Parameter{
				{Name: "age", In: "query", Type: "integer", Required: true},
			},
			url:         "/test?age=25",
			expectError: false,
		},
		{
			name: "Invalid integer parameter",
			params: []Parameter{
				{Name: "age", In: "query", Type: "integer", Required: true},
			},
			url:         "/test?age=not-a-number",
			expectError: true,
			errorCount:  1,
		},
		{
			name: "Valid boolean parameter",
			params: []Parameter{
				{Name: "active", In: "query", Type: "boolean", Required: true},
			},
			url:         "/test?active=true",
			expectError: false,
		},
		{
			name: "Invalid boolean parameter",
			params: []Parameter{
				{Name: "active", In: "query", Type: "boolean", Required: true},
			},
			url:         "/test?active=maybe",
			expectError: true,
			errorCount:  1,
		},
		{
			name: "Valid float parameter",
			params: []Parameter{
				{Name: "price", In: "query", Type: "number", Required: true},
			},
			url:         "/test?price=19.99",
			expectError: false,
		},
		{
			name: "Invalid float parameter",
			params: []Parameter{
				{Name: "price", In: "query", Type: "number", Required: true},
			},
			url:         "/test?price=invalid",
			expectError: true,
			errorCount:  1,
		},
		{
			name: "Optional parameter not provided",
			params: []Parameter{
				{Name: "optional", In: "query", Type: "string", Required: false},
			},
			url:         "/test",
			expectError: false,
		},
		{
			name: "Header parameter",
			params: []Parameter{
				{Name: "X-API-Key", In: "header", Type: "string", Required: true},
			},
			url:         "/test",
			headers:     map[string]string{"X-API-Key": "secret123"},
			expectError: false,
		},
		{
			name: "Missing header parameter",
			params: []Parameter{
				{Name: "X-API-Key", In: "header", Type: "string", Required: true},
			},
			url:         "/test",
			headers:     map[string]string{},
			expectError: true,
			errorCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Get("/test", func(c fiber.Ctx) error {
				err := ValidateParameters(c, tt.params)
				if err != nil {
					return c.Status(400).JSON(err)
				}
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", tt.url, http.NoBody)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to send test request: %v", err)
			}
			defer resp.Body.Close()

			if tt.expectError {
				if resp.StatusCode == 200 {
					t.Errorf("Expected error but got 200 OK")
					return
				}

				var validationErr ValidationErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&validationErr); err != nil {
					t.Errorf("Failed to decode validation error: %v", err)
					return
				}

				if len(validationErr.Errors) != tt.errorCount {
					t.Errorf("Expected %d errors, got %d", tt.errorCount, len(validationErr.Errors))
				}
			} else if resp.StatusCode != 200 {
				t.Errorf("Expected 200 OK but got %d", resp.StatusCode)
			}
		})
	}
}

// TestValidateRequestBody tests request body validation
func TestValidateRequestBody(t *testing.T) {
	tests := []struct {
		schema      interface{}
		body        string
		name        string
		errorField  string
		expectError bool
	}{
		{
			name:        "Valid simple struct",
			body:        `{"name":"John","email":"john@example.com","age":25}`,
			schema:      TestUser{},
			expectError: false,
		},
		{
			name:        "Missing required field",
			body:        `{"name":"John","email":"john@example.com"}`,
			schema:      TestUser{},
			expectError: true,
			errorField:  "age",
		},
		{
			name:        "Invalid field type - string instead of int",
			body:        `{"name":"John","email":"john@example.com","age":"twenty"}`,
			schema:      TestUser{},
			expectError: true,
			errorField:  "age",
		},
		{
			name:        "Invalid field type - number instead of string",
			body:        `{"name":123,"email":"john@example.com","age":25}`,
			schema:      TestUser{},
			expectError: true,
			errorField:  "name",
		},
		{
			name:        "Optional field missing",
			body:        `{"name":"John","is_active":true}`,
			schema:      TestUserWithOptional{},
			expectError: false,
		},
		{
			name:        "Optional field provided",
			body:        `{"name":"John","email":"john@example.com","age":25,"is_active":true}`,
			schema:      TestUserWithOptional{},
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			body:        `{"name":"John",invalid}`,
			schema:      TestUser{},
			expectError: true,
		},
		{
			name:        "Empty body for non-nil schema",
			body:        `{}`,
			schema:      TestUser{},
			expectError: true,
		},
		{
			name:        "Nil schema",
			body:        `{"anything":"goes"}`,
			schema:      nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Post("/test", func(c fiber.Ctx) error {
				err := ValidateRequestBody(c, tt.schema)
				if err != nil {
					return c.Status(400).JSON(err)
				}
				return c.SendString("OK")
			})

			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to send test request: %v", err)
			}
			defer resp.Body.Close()

			if tt.expectError {
				if resp.StatusCode == 200 {
					t.Errorf("Expected error but got 200 OK")
					return
				}

				var validationErr ValidationErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&validationErr); err != nil {
					t.Errorf("Failed to decode validation error: %v", err)
					return
				}

				if tt.errorField != "" && len(validationErr.Errors) > 0 {
					found := false
					for _, e := range validationErr.Errors {
						if e.Field == tt.errorField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error field '%s', but it was not found in errors", tt.errorField)
					}
				}
			} else if resp.StatusCode != 200 {
				t.Errorf("Expected 200 OK but got %d", resp.StatusCode)
			}
		})
	}
}

// TestValidateNestedStruct tests nested struct validation
func TestValidateNestedStruct(t *testing.T) {
	tests := []struct {
		schema      interface{}
		body        string
		name        string
		errorField  string
		expectError bool
	}{
		{
			name: "Valid nested struct",
			body: `{
				"user": {"name":"John","email":"john@example.com","age":25},
				"tags": ["go","api"],
				"enabled": true
			}`,
			schema:      TestNestedStruct{},
			expectError: false,
		},
		{
			name: "Missing nested required field",
			body: `{
				"user": {"name":"John","email":"john@example.com"},
				"tags": ["go","api"],
				"enabled": true
			}`,
			schema:      TestNestedStruct{},
			expectError: true,
			errorField:  "user.age",
		},
		{
			name: "Invalid nested field type",
			body: `{
				"user": {"name":"John","email":"john@example.com","age":"invalid"},
				"tags": ["go","api"],
				"enabled": true
			}`,
			schema:      TestNestedStruct{},
			expectError: true,
			errorField:  "user.age",
		},
		{
			name: "Invalid array element type",
			body: `{
				"user": {"name":"John","email":"john@example.com","age":25},
				"tags": [123, 456],
				"enabled": true
			}`,
			schema:      TestNestedStruct{},
			expectError: true,
			errorField:  "tags[0]",
		},
		{
			name: "Deep nested struct validation",
			body: `{
				"level1": {
					"level2": {
						"level3": {
							"value": "test"
						}
					}
				}
			}`,
			schema:      TestDeepNested{},
			expectError: false,
		},
		{
			name: "Deep nested missing field",
			body: `{
				"level1": {
					"level2": {
						"level3": {}
					}
				}
			}`,
			schema:      TestDeepNested{},
			expectError: true,
			errorField:  "level1.level2.level3.value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Post("/test", func(c fiber.Ctx) error {
				err := ValidateRequestBody(c, tt.schema)
				if err != nil {
					return c.Status(400).JSON(err)
				}
				return c.SendString("OK")
			})

			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to send test request: %v", err)
			}
			defer resp.Body.Close()

			if tt.expectError {
				if resp.StatusCode == 200 {
					t.Errorf("Expected error but got 200 OK")
					return
				}

				var validationErr ValidationErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&validationErr); err != nil {
					t.Errorf("Failed to decode validation error: %v", err)
					return
				}

				if tt.errorField != "" && len(validationErr.Errors) > 0 {
					if validationErr.Errors[0].Field != tt.errorField {
						t.Errorf("Expected error field '%s', got '%s'", tt.errorField, validationErr.Errors[0].Field)
					}
				}
			} else if resp.StatusCode != 200 {
				t.Errorf("Expected 200 OK but got %d", resp.StatusCode)
			}
		})
	}
}

// TestValidateArrayOfStructs tests array of structs validation
func TestValidateArrayOfStructs(t *testing.T) {
	tests := []struct {
		body        string
		name        string
		errorField  string
		expectError bool
	}{
		{
			name: "Valid array of structs",
			body: `{
				"users": [
					{"name":"John","email":"john@example.com","age":25},
					{"name":"Jane","email":"jane@example.com","age":30}
				]
			}`,
			expectError: false,
		},
		{
			name: "Invalid struct in array - missing field",
			body: `{
				"users": [
					{"name":"John","email":"john@example.com","age":25},
					{"name":"Jane","email":"jane@example.com"}
				]
			}`,
			expectError: true,
			errorField:  "users[1].age",
		},
		{
			name: "Invalid struct in array - wrong type",
			body: `{
				"users": [
					{"name":"John","email":"john@example.com","age":25},
					{"name":"Jane","email":"jane@example.com","age":"thirty"}
				]
			}`,
			expectError: true,
			errorField:  "users[1].age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Post("/test", func(c fiber.Ctx) error {
				err := ValidateRequestBody(c, TestArrayOfStructs{})
				if err != nil {
					return c.Status(400).JSON(err)
				}
				return c.SendString("OK")
			})

			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to send test request: %v", err)
			}
			defer resp.Body.Close()

			if tt.expectError {
				if resp.StatusCode == 200 {
					t.Errorf("Expected error but got 200 OK")
					return
				}

				var validationErr ValidationErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&validationErr); err != nil {
					t.Errorf("Failed to decode validation error: %v", err)
					return
				}

				if tt.errorField != "" && len(validationErr.Errors) > 0 {
					if validationErr.Errors[0].Field != tt.errorField {
						t.Errorf("Expected error field '%s', got '%s'", tt.errorField, validationErr.Errors[0].Field)
					}
				}
			} else if resp.StatusCode != 200 {
				t.Errorf("Expected 200 OK but got %d", resp.StatusCode)
			}
		})
	}
}

// TestValidationErrorResponse tests the error interface implementation
func TestValidationErrorResponse(t *testing.T) {
	err := &ValidationErrorResponse{
		ErrorMessage: "Test error",
		Errors: []ValidationError{
			{Field: "test", Message: "Test message", Type: "test_type"},
		},
	}

	if err.Error() != "Test error" {
		t.Errorf("Expected 'Test error', got '%s'", err.Error())
	}
}

// TestValidateParameterType tests parameter type validation
func TestValidateParameterType(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		paramType   string
		expectError bool
	}{
		{"String type", "test", "string", false},
		{"Integer valid", "123", "integer", false},
		{"Integer invalid", "abc", "integer", true},
		{"Float valid", "123.45", "number", false},
		{"Float invalid", "not-a-number", "number", true},
		{"Boolean valid true", "true", "boolean", false},
		{"Boolean valid false", "false", "boolean", false},
		{"Boolean invalid", "maybe", "boolean", true},
		{"Unknown type", "value", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateParameterType(tt.value, tt.paramType)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
