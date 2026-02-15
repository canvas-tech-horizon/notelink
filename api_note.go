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
	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// ApiNote is the main structure for API documentation and routing.
// It integrates with a Fiber application to serve both the API endpoints
// and their documentation.
type ApiNote struct {
	endpoints            map[string]Endpoint
	config               *Config
	app                  *fiber.App
	jwtSecret            string
	middlewares          []fiber.Handler
	jwtMiddlewares       []fiber.Handler
	customAuthMiddleware []fiber.Handler
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
		config:               config,
		endpoints:            make(map[string]Endpoint),
		app:                  app,
		middlewares:          []fiber.Handler{},
		jwtMiddlewares:       []fiber.Handler{},
		customAuthMiddleware: []fiber.Handler{},
		jwtSecret:            jwtSecret,
	}
	app.Get("/api-docs", func(c fiber.Ctx) error {
		// Default to Scalar if DocsUI is empty or explicitly set to "scalar"
		if apiNote.config.DocsUI == "" || apiNote.config.DocsUI == "scalar" {
			return apiNote.ScalarUIHandler()(c)
		} else if apiNote.config.DocsUI == "swagger" {
			return apiNote.SwaggerUIHandler()(c)
		}
		// Fallback to original HTML handler for any other value
		return apiNote.Handler()(c)
	})

	app.Get("/api-docs/metrics", monitor.New(monitor.Config{Title: "Service Metrics Page"}))

	app.Get("/api-docs/indent", func(c fiber.Ctx) error {
		data, err := json.MarshalIndent(app.GetRoutes(true), "", "  ")
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Error marshaling routes")
		}
		return c.Status(http.StatusOK).SendString(string(data))
	})

	// Serve OpenAPI JSON spec at /api-docs/openapi.json
	app.Get("/api-docs/openapi.json", func(c fiber.Ctx) error {
		spec := apiNote.GenerateOpenAPISpec()
		c.Set("Content-Type", "application/json")
		return c.JSON(spec)
	})

	// Serve favicon - try multiple possible locations
	app.Get("/icon.png", func(c fiber.Ctx) error {
		// Try different possible locations for the icon
		iconPaths := []string{
			"./icon.png",     // Current directory
			"../icon.png",    // Parent directory (for examples folder)
			"../../icon.png", // Grandparent directory
		}

		for _, iconPath := range iconPaths {
			if _, err := os.Stat(iconPath); err == nil {
				c.Set("Content-Type", "image/png")
				return c.SendFile(iconPath)
			}
		}

		// If no icon found, return 404
		return c.Status(fiber.StatusNotFound).SendString("Icon not found")
	})

	// Also serve favicon.ico for browsers that look for it by default
	app.Get("/favicon.ico", func(c fiber.Ctx) error {
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
// These middlewares are treated as custom (non-authentication) middleware and will
// not set the AuthRequired flag in endpoint documentation.
//
// Example:
//
//	api.Use(RequestLoggerMiddleware()) // Apply custom logging to all following routes
func (an *ApiNote) Use(middleware ...fiber.Handler) {
	an.middlewares = append(an.middlewares, middleware...)
}

// UseJWT adds JWT authentication middleware to all subsequent routes.
// Routes defined after calling this method will have AuthRequired set to true
// in their documentation, indicating they require authentication.
//
// This is a convenience method that calls JWTMiddleware() and tracks it separately
// from custom middleware, allowing proper documentation of authentication requirements.
//
// Example:
//
//	api.UseJWT() // All routes defined after this will require JWT authentication
func (an *ApiNote) UseJWT() {
	an.jwtMiddlewares = append(an.jwtMiddlewares, an.JWTMiddleware())
}

// UseCustomAuth adds custom authentication middleware to all subsequent routes.
// Routes defined after calling this method will have AuthRequired set to true
// in their documentation, indicating they require authentication.
//
// This method allows you to use your own authentication logic instead of the
// built-in JWT middleware while still properly documenting that routes require authentication.
//
// Example:
//
//	api.UseCustomAuth(MyCustomAuthMiddleware()) // All routes defined after this will require custom authentication
func (an *ApiNote) UseCustomAuth(middleware ...fiber.Handler) {
	an.customAuthMiddleware = append(an.customAuthMiddleware, middleware...)
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
func (an *ApiNote) DocumentedRoute(input *DocumentedRouteInput) error {
	// Validate required fields
	if input.Method == "" || input.Path == "" {
		return fmt.Errorf("method and path are required")
	}
	if input.Handler == nil {
		return fmt.Errorf("handler is required")
	}

	key := input.Method + " " + input.Path
	endpoint := Endpoint{
		Method:      input.Method,
		Path:        an.config.BasePath + input.Path,
		Description: input.Description,
		Responses:   input.Responses,
		Parameters:  input.Params,
	}

	// Set AuthRequired based on explicit input or JWT middleware presence
	if input.AuthRequired != nil {
		endpoint.AuthRequired = *input.AuthRequired
	} else {
		// Auto-detect: true if JWT middleware or custom auth middleware is active
		endpoint.AuthRequired = len(an.jwtMiddlewares) > 0 || len(an.customAuthMiddleware) > 0
	}

	if input.SchemasRequest != nil {
		endpoint.RequestSchema = input.SchemasRequest
	}
	if input.SchemasResponse != nil {
		endpoint.ResponseSchema = input.SchemasResponse
	}

	an.endpoints[key] = endpoint

	// Combine authentication middlewares (JWT or custom), custom middlewares, then add the handler
	handlers := []any{}
	if endpoint.AuthRequired {
		// Add JWT middlewares if present
		for _, h := range an.jwtMiddlewares {
			handlers = append(handlers, h)
		}
		// Add custom auth middlewares if present
		for _, h := range an.customAuthMiddleware {
			handlers = append(handlers, h)
		}
	}

	// Add validation middleware if validation is needed
	// Validation is enabled by default when parameters or request schema are defined
	if len(endpoint.Parameters) > 0 || endpoint.RequestSchema != nil {
		validationHandler := func(c fiber.Ctx) error {
			// Validate parameters
			if len(endpoint.Parameters) > 0 {
				if err := ValidateParameters(c, endpoint.Parameters); err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(err)
				}
			}

			// Validate request body for POST/PUT/PATCH
			if endpoint.RequestSchema != nil &&
				(endpoint.Method == "POST" || endpoint.Method == "PUT" || endpoint.Method == "PATCH") {
				if err := ValidateRequestBody(c, endpoint.RequestSchema); err != nil {
					return c.Status(fiber.StatusBadRequest).JSON(err)
				}
			}

			return c.Next()
		}
		handlers = append(handlers, validationHandler)
	}

	// Add custom non-auth middlewares
	for _, h := range an.middlewares {
		handlers = append(handlers, h)
	}
	// Add the route handler
	handlers = append(handlers, input.Handler)

	// Ensure we have at least one handler
	if len(handlers) == 0 {
		return fmt.Errorf("at least one handler is required")
	}

	path := an.config.BasePath + input.Path
	// Get first handler and rest as varargs for v3 API
	firstHandler := handlers[0]
	restHandlers := []any{}
	if len(handlers) > 1 {
		restHandlers = handlers[1:]
	}

	switch strings.ToUpper(input.Method) {
	case "GET":
		an.app.Get(path, firstHandler, restHandlers...)
	case "POST":
		an.app.Post(path, firstHandler, restHandlers...)
	case "PUT":
		an.app.Put(path, firstHandler, restHandlers...)
	case "DELETE":
		an.app.Delete(path, firstHandler, restHandlers...)
	case "PATCH":
		an.app.Patch(path, firstHandler, restHandlers...)
	case "HEAD":
		an.app.Head(path, firstHandler, restHandlers...)
	case "CONNECT":
		an.app.Connect(path, firstHandler, restHandlers...)
	case "OPTIONS":
		an.app.Options(path, firstHandler, restHandlers...)
	case "TRACE":
		an.app.Trace(path, firstHandler, restHandlers...)
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
	return func(c fiber.Ctx) error {
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
	return func(c fiber.Ctx) error {
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
