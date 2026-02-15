package notelink

import "github.com/gofiber/fiber/v3"

// Config holds the API documentation configuration
type Config struct {
	Title                string
	Description          string
	Version              string
	Host                 string
	BasePath             string
	AuthToken            string // Optional authorization token (e.g., Bearer token)
	DocsUI               string // UI to use for /api-docs endpoint: "scalar" (default) or "swagger"
	EnableValidation     bool   // Enable server-side validation (default: true)
	StrictTypeValidation bool   // Strict type checking vs coercion (default: false)
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string
	In          string // "query", "path", "header"
	Type        string // e.g., "string", "number", "boolean"
	Description string
	Required    bool
}

// Endpoint represents a single API endpoint with schema and parameters
type Endpoint struct {
	Method         string
	Path           string
	Description    string
	Responses      map[string]string
	RequestSchema  interface{}
	ResponseSchema interface{}
	Parameters     []Parameter
	AuthRequired   bool // Indicates if authorization is required
}

// DocumentedRouteInput represents the input for registering a documented route
type DocumentedRouteInput struct {
	Params          []Parameter       `json:"params"`
	SchemasRequest  interface{}       `json:"schemasRequest"`
	SchemasResponse interface{}       `json:"schemasResponse"`
	Responses       map[string]string `json:"responses"`
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	Description     string            `json:"description"`
	Handler         fiber.Handler     `json:"handler"`
	AuthRequired    *bool             `json:"authRequired"` // Optional: explicitly set if auth is required. If nil, auto-detected based on JWT middleware usage
}
