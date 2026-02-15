package notelink

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// OpenAPI 3.1 root structure
type OpenAPISpec struct {
	OpenAPI    string                `json:"openapi"`
	Info       OpenAPIInfo           `json:"info"`
	Servers    []OpenAPIServer       `json:"servers,omitempty"`
	Paths      map[string]PathItem   `json:"paths"`
	Components *Components           `json:"components,omitempty"`
	Security   []map[string][]string `json:"security,omitempty"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Trace   *Operation `json:"trace,omitempty"`
}

type Operation struct {
	OperationID string                `json:"operationId"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Parameters  []ParameterSpec       `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
}

type ParameterSpec struct {
	Schema      *JSONSchema `json:"schema"`
	Name        string      `json:"name"`
	In          string      `json:"in"` // "query", "path", "header", "cookie"
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
}

type RequestBody struct {
	Content     map[string]MediaType `json:"content"`
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required,omitempty"`
}

type Response struct {
	Content     map[string]MediaType `json:"content,omitempty"`
	Description string               `json:"description"`
}

type MediaType struct {
	Schema  *JSONSchema `json:"schema,omitempty"`
	Example interface{} `json:"example,omitempty"`
}

type Components struct {
	Schemas         map[string]*JSONSchema    `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"` // "http", "apiKey", "oauth2", "openIdConnect"
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
}

// JSONSchema represents JSON Schema (compatible with OpenAPI 3.1)
type JSONSchema struct {
	AdditionalProperties interface{}            `json:"additionalProperties,omitempty"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`
	Items                *JSONSchema            `json:"items,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Ref                  string                 `json:"$ref,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Nullable             bool                   `json:"nullable,omitempty"`
}

// GenerateOpenAPISpec creates an OpenAPI 3.1 specification from registered endpoints
func (an *ApiNote) GenerateOpenAPISpec() *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: OpenAPIInfo{
			Title:       an.config.Title,
			Description: an.config.Description,
			Version:     an.config.Version,
		},
		Servers: []OpenAPIServer{
			{
				URL:         "http://" + an.config.Host + an.config.BasePath,
				Description: "API Server",
			},
		},
		Paths: make(map[string]PathItem),
		Components: &Components{
			Schemas:         make(map[string]*JSONSchema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	// Check if any endpoint requires authentication
	hasAuth := false
	for _, endpoint := range an.endpoints {
		if endpoint.AuthRequired {
			hasAuth = true
			break
		}
	}

	// Add JWT Bearer security scheme if authentication is used
	if hasAuth {
		spec.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT Authorization header using the Bearer scheme",
		}
	}

	// Process each endpoint
	for _, endpoint := range an.endpoints {
		pathItem, ok := spec.Paths[endpoint.Path]
		if !ok {
			pathItem = PathItem{}
		}

		operation := an.endpointToOperation(&endpoint, spec.Components.Schemas)

		// Assign operation to the correct HTTP method
		switch strings.ToUpper(endpoint.Method) {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		case "HEAD":
			pathItem.Head = operation
		case "OPTIONS":
			pathItem.Options = operation
		case "TRACE":
			pathItem.Trace = operation
		}

		spec.Paths[endpoint.Path] = pathItem
	}

	return spec
}

// endpointToOperation converts an Endpoint to an OpenAPI Operation
func (an *ApiNote) endpointToOperation(endpoint *Endpoint, componentSchemas map[string]*JSONSchema) *Operation {
	// Generate operation ID from method and path
	operationID := generateOperationID(endpoint.Method, endpoint.Path)

	operation := &Operation{
		OperationID: operationID,
		Summary:     endpoint.Description,
		Description: endpoint.Description,
		Parameters:  []ParameterSpec{},
		Responses:   make(map[string]Response),
	}

	// Extract tags from path (e.g., "/api/v1/users" -> ["users"])
	tags := extractTagsFromPath(endpoint.Path)
	if len(tags) > 0 {
		operation.Tags = tags
	}

	// Add security requirement if endpoint requires authentication
	if endpoint.AuthRequired {
		operation.Security = []map[string][]string{
			{"bearerAuth": []string{}},
		}
	}

	// Convert parameters
	for _, param := range endpoint.Parameters {
		paramSchema := parameterTypeToJSONSchema(param.Type)
		paramSpec := ParameterSpec{
			Name:        param.Name,
			In:          param.In,
			Description: param.Description,
			Required:    param.Required,
			Schema:      paramSchema,
		}
		operation.Parameters = append(operation.Parameters, paramSpec)
	}

	// Add request body if RequestSchema exists
	if endpoint.RequestSchema != nil {
		schema, nestedSchemas := generateJSONSchema("RequestBody", endpoint.RequestSchema)

		// Add nested schemas to components
		for name, nestedSchema := range nestedSchemas {
			if _, exists := componentSchemas[name]; !exists {
				componentSchemas[name] = nestedSchema
			}
		}

		// Generate example from schema
		exampleJSON, err := generateJSONTemplate(endpoint.RequestSchema)
		if err == nil {
			var exampleData interface{}
			if err := json.Unmarshal([]byte(exampleJSON), &exampleData); err == nil {
				operation.RequestBody = &RequestBody{
					Required: true,
					Content: map[string]MediaType{
						"application/json": {
							Schema:  schema,
							Example: exampleData,
						},
					},
				}
			}
		}
	}

	// Add responses
	for statusCode, description := range endpoint.Responses {
		response := Response{
			Description: description,
		}

		// Add response schema for successful responses
		if statusCode == "200" || statusCode == "201" {
			if endpoint.ResponseSchema != nil {
				schema, nestedSchemas := generateJSONSchema("ResponseBody", endpoint.ResponseSchema)

				// Add nested schemas to components
				for name, nestedSchema := range nestedSchemas {
					if _, exists := componentSchemas[name]; !exists {
						componentSchemas[name] = nestedSchema
					}
				}

				// Generate example from schema
				exampleJSON, err := generateJSONTemplate(endpoint.ResponseSchema)
				if err == nil {
					var exampleData interface{}
					if err := json.Unmarshal([]byte(exampleJSON), &exampleData); err == nil {
						response.Content = map[string]MediaType{
							"application/json": {
								Schema:  schema,
								Example: exampleData,
							},
						}
					}
				}
			}
		}

		operation.Responses[statusCode] = response
	}

	// Ensure at least a default response exists
	if len(operation.Responses) == 0 {
		operation.Responses["200"] = Response{
			Description: "Successful response",
		}
	}

	return operation
}

// generateJSONSchema converts a Go type to JSON Schema format
func generateJSONSchema(name string, schema interface{}) (mainSchema *JSONSchema, componentSchemas map[string]*JSONSchema) {
	if schema == nil {
		return &JSONSchema{Type: "object"}, nil
	}

	typ := reflect.TypeOf(schema)
	if typ == nil {
		return &JSONSchema{Type: "object"}, nil
	}

	// Handle pointers and slices at the top level
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	isArray := false
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		isArray = true
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
	}

	if typ.Kind() != reflect.Struct {
		return goTypeToJSONSchema(typ), nil
	}

	// Generate schemas for all nested structs
	componentSchemas = make(map[string]*JSONSchema)
	collectComponentSchemas(typ, componentSchemas)

	// Generate the main schema
	mainSchema = structToJSONSchema(typ, name, componentSchemas)

	if isArray {
		return &JSONSchema{
			Type:  "array",
			Items: mainSchema,
		}, componentSchemas
	}

	return mainSchema, componentSchemas
}

// collectComponentSchemas recursively collects all nested struct schemas
func collectComponentSchemas(typ reflect.Type, schemas map[string]*JSONSchema) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
	}

	if typ.Kind() != reflect.Struct || typ.Name() == "" {
		return
	}

	// Skip if already processed
	if _, exists := schemas[typ.Name()]; exists {
		return
	}

	// Special case for time.Time
	if typ == reflect.TypeOf(time.Time{}) {
		return
	}

	// Process all fields to find nested structs
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
		}

		if fieldType.Kind() == reflect.Struct && fieldType.Name() != "" && fieldType != reflect.TypeOf(time.Time{}) {
			collectComponentSchemas(fieldType, schemas)
		}
	}

	// Add this struct to schemas
	schemas[typ.Name()] = structToJSONSchema(typ, typ.Name(), schemas)
}

// structToJSONSchema converts a struct type to JSON Schema
func structToJSONSchema(typ reflect.Type, name string, componentSchemas map[string]*JSONSchema) *JSONSchema {
	schema := &JSONSchema{
		Type:       "object",
		Title:      name,
		Properties: make(map[string]*JSONSchema),
		Required:   []string{},
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if !field.IsExported() {
			continue
		}

		fieldName := getJSONFieldName(&field)
		if fieldName == "-" {
			continue
		}

		fieldSchema := fieldToJSONSchema(field.Type, field.Name, componentSchemas)
		schema.Properties[fieldName] = fieldSchema

		// Check if field is required (not a pointer and no omitempty tag)
		jsonTag := field.Tag.Get("json")
		isOmitEmpty := strings.Contains(jsonTag, "omitempty")
		isPointer := field.Type.Kind() == reflect.Ptr

		if !isOmitEmpty && !isPointer {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	// Remove required array if empty
	if len(schema.Required) == 0 {
		schema.Required = nil
	}

	return schema
}

// fieldToJSONSchema converts a field type to JSON Schema
func fieldToJSONSchema(t reflect.Type, fieldName string, componentSchemas map[string]*JSONSchema) *JSONSchema {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		schema := fieldToJSONSchema(t.Elem(), fieldName, componentSchemas)
		schema.Nullable = true
		return schema
	}

	// Handle slices
	if t.Kind() == reflect.Slice {
		return &JSONSchema{
			Type:  "array",
			Items: fieldToJSONSchema(t.Elem(), fieldName, componentSchemas),
		}
	}

	// Handle structs
	if t.Kind() == reflect.Struct {
		// Special case for time.Time
		if t == reflect.TypeOf(time.Time{}) {
			return &JSONSchema{
				Type:   "string",
				Format: "date-time",
			}
		}

		// Reference to component schema if it has a name
		if t.Name() != "" {
			return &JSONSchema{
				Ref: "#/components/schemas/" + t.Name(),
			}
		}

		// Anonymous struct - inline it
		return structToJSONSchema(t, "", componentSchemas)
	}

	return goTypeToJSONSchema(t)
}

// goTypeToJSONSchema maps Go primitive types to JSON Schema types
func goTypeToJSONSchema(t reflect.Type) *JSONSchema {
	switch t.Kind() {
	case reflect.String:
		return &JSONSchema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return &JSONSchema{Type: "integer", Format: "int32"}
	case reflect.Int64:
		return &JSONSchema{Type: "integer", Format: "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		minVal := 0.0
		return &JSONSchema{Type: "integer", Format: "int32", Minimum: &minVal}
	case reflect.Uint64:
		minVal := 0.0
		return &JSONSchema{Type: "integer", Format: "int64", Minimum: &minVal}
	case reflect.Float32:
		return &JSONSchema{Type: "number", Format: "float"}
	case reflect.Float64:
		return &JSONSchema{Type: "number", Format: "double"}
	case reflect.Bool:
		return &JSONSchema{Type: "boolean"}
	default:
		return &JSONSchema{} // Empty schema for unknown types
	}
}

// parameterTypeToJSONSchema converts Parameter.Type string to JSON Schema
func parameterTypeToJSONSchema(paramType string) *JSONSchema {
	switch strings.ToLower(paramType) {
	case "string":
		return &JSONSchema{Type: "string"}
	case "number", "float", "double":
		return &JSONSchema{Type: "number"}
	case "integer", "int":
		return &JSONSchema{Type: "integer"}
	case "boolean", "bool":
		return &JSONSchema{Type: "boolean"}
	default:
		return &JSONSchema{Type: "string"}
	}
}

// generateOperationID creates a unique operation ID from method and path
// Example: GET /api/v1/users/:id -> getUsersById
func generateOperationID(method, path string) string {
	// Clean path and split into segments
	cleanPath := strings.Trim(path, "/")
	segments := strings.Split(cleanPath, "/")

	// Build operation ID
	parts := make([]string, 0, len(segments)+1)
	parts = append(parts, strings.ToLower(method))

	for _, segment := range segments {
		// Skip common prefixes like "api", "v1", "v2", etc.
		if segment == "api" || (strings.HasPrefix(segment, "v") && len(segment) <= 3) {
			continue
		}

		// Handle path parameters like :id or {id}
		switch {
		case strings.HasPrefix(segment, ":"):
			segment = "By" + toTitle(segment[1:])
		case strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}"):
			segment = "By" + toTitle(segment[1:len(segment)-1])
		default:
			segment = toTitle(segment)
		}

		parts = append(parts, segment)
	}

	return strings.Join(parts, "")
}

// toTitle converts the first character of a string to uppercase
func toTitle(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// extractTagsFromPath extracts resource tags from the path
// Example: /api/v1/users/:id -> ["users"]
func extractTagsFromPath(path string) []string {
	segments := strings.Split(strings.Trim(path, "/"), "/")

	for _, segment := range segments {
		// Find the first non-version, non-api segment that's not a parameter
		if segment != "" && segment != "api" &&
			!strings.HasPrefix(segment, "v") &&
			!strings.HasPrefix(segment, ":") &&
			!strings.HasPrefix(segment, "{") {
			return []string{segment}
		}
	}

	return nil
}

// ExportOpenAPIToFile exports the OpenAPI specification to a JSON file
func (an *ApiNote) ExportOpenAPIToFile(filepath string) error {
	spec := an.GenerateOpenAPISpec()

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	err = os.WriteFile(filepath, data, 0o600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
