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
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
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
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})
	apiNote := &ApiNote{
		config:      config,
		endpoints:   make(map[string]Endpoint),
		app:         app,
		middlewares: []fiber.Handler{},
		jwtSecret:   jwtSecret,
	}
	app.Get("/api-docs", apiNote.Handler())

	app.Get("/api-docs/metrics", monitor.New(monitor.Config{Title: "Service Metrics Page"}))

	app.Get("/api-docs/indent", func(c *fiber.Ctx) error {
		data, _ := json.MarshalIndent(app.GetRoutes(true), "", "  ")
		return c.Status(http.StatusOK).SendString(string(data))
	})

	// Serve favicon - try multiple possible locations
	app.Get("/icon.png", func(c *fiber.Ctx) error {
		// Try different possible locations for the icon
		iconPaths := []string{
			"./icon.png",     // Current directory
			"../icon.png",    // Parent directory (for examples folder)
			"../../icon.png", // Grandparent directory
		}

		for _, iconPath := range iconPaths {
			if _, err := os.Stat(iconPath); err == nil {
				c.Set("Content-Type", "image/png")
				c.Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
				return c.SendFile(iconPath)
			}
		}

		// If no icon found, return 404
		return c.Status(fiber.StatusNotFound).SendString("Icon not found")
	})

	// Also serve favicon.ico for browsers that look for it by default
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		// Try different possible locations for the icon
		iconPaths := []string{
			"./icon.png",     // Current directory
			"../icon.png",    // Parent directory (for examples folder)
			"../../icon.png", // Grandparent directory
		}

		for _, iconPath := range iconPaths {
			if _, err := os.Stat(iconPath); err == nil {
				c.Set("Content-Type", "image/x-icon")
				return c.SendFile(iconPath)
			}
		}

		// If no icon found, return 404
		return c.Status(fiber.StatusNotFound).SendString("Icon not found")
	})

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
// It accepts a DocumentedRouteInput object containing the route details and
// processes it to add the route to the Fiber app and store endpoint details for documentation.
//
// Parameters:
//   - input: DocumentedRouteInput object containing method, path, description, responses,
//     handler, params, schemasRequest, and schemasResponse
//
// Example usage:
//
//	api.DocumentedRoute(notelink.DocumentedRouteInput{
//	    Method:      "POST",
//	    Path:        "/v3/users",
//	    Description: "Create a new user (Authenticated)",
//	    Responses: map[string]string{
//	        "201": "User created",
//	        "400": "Invalid input",
//	        "401": "Unauthorized",
//	    },
//	    Handler: handlerFunc,
//	    Params:  []notelink.Parameter{},
//	    SchemasRequest:  CreateUserRequest{},
//	    SchemasResponse: UserResponse{},
//	})
func (an *ApiNote) DocumentedRoute(input DocumentedRouteInput) error {
	// Validate required fields
	if input.Method == "" || input.Path == "" {
		return fmt.Errorf("method and path are required")
	}
	if input.Handler == nil {
		return fmt.Errorf("handler is required")
	}

	key := input.Method + " " + input.Path
	endpoint := Endpoint{
		Method:       input.Method,
		Path:         an.config.BasePath + input.Path,
		Description:  input.Description,
		Responses:    input.Responses,
		Parameters:   input.Params,
		AuthRequired: len(an.middlewares) > 0,
	}

	if input.SchemasRequest != nil {
		endpoint.RequestSchema = input.SchemasRequest
	}
	if input.SchemasResponse != nil {
		endpoint.ResponseSchema = input.SchemasResponse
	}

	an.endpoints[key] = endpoint

	handlers := append(an.middlewares, input.Handler)
	switch strings.ToUpper(input.Method) {
	case "GET":
		an.app.Get(an.config.BasePath+input.Path, handlers...)
	case "POST":
		an.app.Post(an.config.BasePath+input.Path, handlers...)
	case "PUT":
		an.app.Put(an.config.BasePath+input.Path, handlers...)
	case "DELETE":
		an.app.Delete(an.config.BasePath+input.Path, handlers...)
	case "PATCH":
		an.app.Patch(an.config.BasePath+input.Path, handlers...)
	case "HEAD":
		an.app.Head(an.config.BasePath+input.Path, handlers...)
	case "CONNECT":
		an.app.Connect(an.config.BasePath+input.Path, handlers...)
	case "OPTIONS":
		an.app.Options(an.config.BasePath+input.Path, handlers...)
	case "TRACE":
		an.app.Trace(an.config.BasePath+input.Path, handlers...)
	default:
		return fmt.Errorf("unsupported HTTP method: %s", input.Method)
	}

	return nil
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
