// Package notelink provides a framework for generating API documentation
// and integrating it with a Fiber web server. It allows developers to define
// endpoints with parameters, schemas, and authentication, and automatically
// generates an interactive HTML documentation page.
//
// The package supports:
//   - Endpoint documentation with methods, paths, descriptions, and responses
//   - Parameter definitions (query, path, header)
//   - Automatic TypeScript schema generation from Go structs
//   - JWT authentication middleware
//   - Interactive API testing via a generated HTML interface
//
// To use this package, create an ApiNote instance with a configuration and
// JWT secret, then define routes using the DocumentedRoute method. The
// documentation is served at "/api-docs" by default.
//
// Example:
//
//	config := notelink.Config{Title: "My API", Host: "localhost:8080"}
//	api := notelink.NewApiNote(&config, "secret-key")
//	api.DocumentedRoute("GET", "/api/v1/hello", "Say hello", map[string]string{"200": "Success"}, handler, nil)
//	api.Listen()
package notelink

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ApiNote is the main structure for API documentation and routing.
// It integrates with a Fiber application to serve both the API endpoints
// and their documentation.
type ApiNote struct {
	config      *Config             // Configuration for the API documentation
	endpoints   map[string]Endpoint // Registered endpoints with their details
	app         *fiber.App          // Underlying Fiber application
	middlewares []fiber.Handler     // Middleware stack applied to routes
	jwtSecret   string              // Secret key for JWT signing and verification
}

// NewApiNote creates a new ApiNote instance with the provided configuration and JWT secret.
// It initializes a Fiber application and sets up the "/api-docs" route for documentation.
//
// The config parameter defines the API's title, host, and other metadata.
// The jwtSecret is used for JWT authentication middleware.
//
// Returns a pointer to the initialized ApiNote.
func NewApiNote(config *Config, jwtSecret string) *ApiNote {
	app := fiber.New()
	apiNote := &ApiNote{
		config:      config,
		endpoints:   make(map[string]Endpoint),
		app:         app,
		middlewares: []fiber.Handler{},
		jwtSecret:   jwtSecret,
	}
	app.Get("/api-docs", apiNote.Handler())
	return apiNote
}

// Use adds one or more middleware handlers to be applied to all subsequent routes.
// Middleware is executed in the order it is added.
//
// Example:
//
//	api.Use(api.JWTMiddleware()) // Apply JWT authentication to all following routes
func (an *ApiNote) Use(middleware ...fiber.Handler) {
	an.middlewares = append(an.middlewares, middleware...)
}

// DocumentedRoute registers an API endpoint with its documentation and handler.
// It adds the route to the Fiber app and stores the endpoint details for documentation.
//
// Parameters:
//   - method: HTTP method (e.g., "GET", "POST")
//   - path: Endpoint path (e.g., "/api/v1/users/:id")
//   - description: Brief description of the endpoint
//   - responses: Map of status codes to response descriptions
//   - handler: Fiber handler function for the endpoint
//   - params: List of parameters (query, path, or header)
//   - schemas: Optional request and response schemas (first is request, second is response)
//
// The endpoint is marked as requiring authentication if any middleware is active.
func (an *ApiNote) DocumentedRoute(
	method, path, description string,
	responses map[string]string,
	handler fiber.Handler,
	params []Parameter,
	schemas ...interface{},
) {
	key := method + " " + path
	endpoint := Endpoint{
		Method:       method,
		Path:         an.config.BasePath + path,
		Description:  description,
		Responses:    responses,
		Parameters:   params,
		AuthRequired: len(an.middlewares) > 0,
	}

	if len(schemas) > 0 && schemas[0] != nil {
		endpoint.RequestSchema = schemas[0]
	}
	if len(schemas) > 1 && schemas[1] != nil {
		endpoint.ResponseSchema = schemas[1]
	}

	an.endpoints[key] = endpoint

	handlers := append(an.middlewares, handler)
	switch method {
	case "GET":
		an.app.Get(an.config.BasePath+path, handlers...)
	case "POST":
		an.app.Post(an.config.BasePath+path, handlers...)
	case "PUT":
		an.app.Put(an.config.BasePath+path, handlers...)
	case "DELETE":
		an.app.Delete(an.config.BasePath+path, handlers...)
	case "PATCH":
		an.app.Patch(an.config.BasePath+path, handlers...)
	}
}

// Fiber returns the underlying *fiber.App instance used by the ApiNote.
//
// This allows external packages or components to directly access and
// configure the Fiber application (e.g., for adding routes, middleware, etc.).
func (an *ApiNote) Fiber() *fiber.App {
	return an.app
}

// Handler returns a Fiber handler that serves the API documentation as HTML.
// The documentation is generated dynamically based on registered endpoints.
//
// The returned handler sets the Content-Type to "text/html" and responds with status 200.
func (an *ApiNote) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		html := an.generateHTML()
		c.Set("Content-Type", "text/html")
		return c.Status(http.StatusOK).SendString(html)
	}
}

// JWTMiddleware returns a Fiber middleware handler that validates JWT tokens.
// It checks the "Authorization" header for a "Bearer" token and verifies it
// using the configured jwtSecret.
//
// If the token is valid, it sets the "user_id" in the context from the token's "sub" claim.
// If invalid or missing, it returns a 401 Unauthorized response.
//
// Example usage:
//
//	api.Use(api.JWTMiddleware())
func (an *ApiNote) JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header required"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Authorization header format"})
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(an.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("user_id", claims["sub"])
		}

		return c.Next()
	}
}

// Listen starts the Fiber server on the port specified in Config.Host.
// The Host field should be in the format "host:port" (e.g., "localhost:8080").
// If no port is specified, it defaults to ":8080".
//
// Returns an error if the server fails to start.
func (an *ApiNote) Listen() error {
	hostParts := strings.Split(an.config.Host, ":")
	port := ":8080" // Default port
	if len(hostParts) > 1 {
		port = ":" + hostParts[1]
	}
	return an.app.Listen(port)
}
