package notelink

import "github.com/gofiber/fiber/v2"

// Config holds the API documentation configuration
type Config struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
	AuthToken   string // Optional authorization token (e.g., Bearer token)
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
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	Description     string            `json:"description"`
	Responses       map[string]string `json:"responses"`
	Handler         fiber.Handler     `json:"handler"`
	Params          []Parameter       `json:"params"`
	SchemasRequest  interface{}       `json:"schemasRequest"`
	SchemasResponse interface{}       `json:"schemasResponse"`
}
