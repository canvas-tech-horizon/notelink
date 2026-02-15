package notelink

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
}

// ValidationErrorResponse represents the validation error response
type ValidationErrorResponse struct {
	ErrorMessage string            `json:"error"`
	Errors       []ValidationError `json:"errors,omitempty"`
}

// Error implements the error interface
func (v *ValidationErrorResponse) Error() string {
	return v.ErrorMessage
}

// ValidateParameters validates path/query/header parameters
func ValidateParameters(c fiber.Ctx, params []Parameter) error {
	var errors []ValidationError

	for _, param := range params {
		value, exists := getParameterValue(c, param)

		// Check if required parameter is missing
		if param.Required && (!exists || value == "") {
			errors = append(errors, ValidationError{
				Field:   param.Name,
				Message: fmt.Sprintf("Required parameter '%s' is missing", param.Name),
				Type:    "required",
			})
			continue
		}

		// Skip validation if parameter is not provided and not required
		if !exists || value == "" {
			continue
		}

		// Validate parameter type
		if _, err := validateParameterType(value, param.Type); err != nil {
			errors = append(errors, ValidationError{
				Field:   param.Name,
				Message: fmt.Sprintf("Parameter '%s' must be of type %s: %v", param.Name, param.Type, err),
				Type:    "type_error",
			})
		}
	}

	if len(errors) > 0 {
		return &ValidationErrorResponse{
			ErrorMessage: "Parameter validation failed",
			Errors:       errors,
		}
	}

	return nil
}

// ValidateRequestBody validates request body against schema
func ValidateRequestBody(c fiber.Ctx, schema interface{}) error {
	if schema == nil {
		return nil
	}

	// Get request body
	var body map[string]interface{}
	if err := c.Bind().Body(&body); err != nil {
		return &ValidationErrorResponse{
			ErrorMessage: "Invalid JSON body",
			Errors: []ValidationError{{
				Field:   "body",
				Message: err.Error(),
				Type:    "parse_error",
			}},
		}
	}

	// Validate against schema using reflection
	schemaType := reflect.TypeOf(schema)
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}

	// Handle array schemas
	if schemaType.Kind() == reflect.Slice {
		// For array schemas, we don't validate structure
		// Just ensure body can be parsed
		return nil
	}

	errors := validateStruct(body, schemaType)
	if len(errors) > 0 {
		return &ValidationErrorResponse{
			ErrorMessage: "Request body validation failed",
			Errors:       errors,
		}
	}

	return nil
}

// getParameterValue extracts parameter value from request based on parameter location
func getParameterValue(c fiber.Ctx, param Parameter) (string, bool) {
	switch param.In {
	case "path":
		value := c.Params(param.Name)
		return value, value != ""
	case "query":
		value := c.Query(param.Name)
		return value, value != ""
	case "header":
		value := c.Get(param.Name)
		return value, value != ""
	default:
		return "", false
	}
}

// validateParameterType validates and converts parameter value to the expected type
func validateParameterType(value string, paramType string) (interface{}, error) {
	switch strings.ToLower(paramType) {
	case "string":
		return value, nil
	case "number", "float", "double":
		return strconv.ParseFloat(value, 64)
	case "integer", "int":
		return strconv.Atoi(value)
	case "boolean", "bool":
		return strconv.ParseBool(value)
	default:
		// Unknown type, pass through as string
		return value, nil
	}
}

// validateStruct validates a map against a struct type
func validateStruct(data map[string]interface{}, schemaType reflect.Type) []ValidationError {
	var errors []ValidationError

	// Handle non-struct types
	if schemaType.Kind() != reflect.Struct {
		return errors
	}

	for i := 0; i < schemaType.NumField(); i++ {
		field := schemaType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON field name
		jsonName := getJSONFieldName(field)
		if jsonName == "-" {
			continue
		}

		// Check if field is required (not pointer, no omitempty)
		jsonTag := field.Tag.Get("json")
		isOmitEmpty := strings.Contains(jsonTag, "omitempty")
		isPointer := field.Type.Kind() == reflect.Ptr
		isRequired := !isOmitEmpty && !isPointer

		value, exists := data[jsonName]

		// Check required fields
		if isRequired && (!exists || value == nil) {
			errors = append(errors, ValidationError{
				Field:   jsonName,
				Message: fmt.Sprintf("Required field '%s' is missing", jsonName),
				Type:    "required",
			})
			continue
		}

		// Validate field type if value exists
		if exists && value != nil {
			if err := validateFieldType(value, field.Type, jsonName); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	return errors
}

// validateFieldType validates the type of a field value
func validateFieldType(value interface{}, expectedType reflect.Type, fieldName string) *ValidationError {
	// Handle pointers
	if expectedType.Kind() == reflect.Ptr {
		expectedType = expectedType.Elem()
	}

	actualValue := reflect.ValueOf(value)
	if !actualValue.IsValid() {
		return nil // nil value is okay for optional fields
	}

	switch expectedType.Kind() {
	case reflect.String:
		if actualValue.Kind() != reflect.String {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be a string", fieldName),
				Type:    "type_error",
			}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// JSON unmarshals numbers as float64
		if actualValue.Kind() != reflect.Float64 {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be a number", fieldName),
				Type:    "type_error",
			}
		}
		// Check if it's an integer value
		floatVal := actualValue.Float()
		if floatVal != float64(int64(floatVal)) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be an integer", fieldName),
				Type:    "type_error",
			}
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// JSON unmarshals numbers as float64
		if actualValue.Kind() != reflect.Float64 {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be a number", fieldName),
				Type:    "type_error",
			}
		}
		// Check if it's a non-negative integer value
		floatVal := actualValue.Float()
		if floatVal < 0 || floatVal != float64(int64(floatVal)) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be a non-negative integer", fieldName),
				Type:    "type_error",
			}
		}

	case reflect.Float32, reflect.Float64:
		if actualValue.Kind() != reflect.Float64 && actualValue.Kind() != reflect.Int && actualValue.Kind() != reflect.Int64 {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be a number", fieldName),
				Type:    "type_error",
			}
		}

	case reflect.Bool:
		if actualValue.Kind() != reflect.Bool {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be a boolean", fieldName),
				Type:    "type_error",
			}
		}

	case reflect.Slice, reflect.Array:
		if actualValue.Kind() != reflect.Slice && actualValue.Kind() != reflect.Array {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be an array", fieldName),
				Type:    "type_error",
			}
		}
		// TODO: Validate array elements recursively

	case reflect.Map:
		if actualValue.Kind() != reflect.Map {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be an object", fieldName),
				Type:    "type_error",
			}
		}

	case reflect.Struct:
		// Nested struct should be a map in JSON
		if actualValue.Kind() != reflect.Map {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("Field '%s' must be an object", fieldName),
				Type:    "type_error",
			}
		}
		// TODO: Validate nested struct recursively
	}

	return nil
}
