package notelink

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ApiNote is the main structure for our documentation
type ApiNote struct {
	config      *Config
	endpoints   map[string]Endpoint
	router      *gin.Engine
	middlewares []gin.HandlerFunc
	jwtSecret   string // Secret key for JWT signing and verification
}

// NewApiNote creates a new ApiNote instance integrated with Gin
func NewApiNote(router *gin.Engine, config *Config, jwtSecret string) *ApiNote {
	apiNote := &ApiNote{
		config:      config,
		endpoints:   make(map[string]Endpoint),
		router:      router,
		middlewares: []gin.HandlerFunc{},
		jwtSecret:   jwtSecret,
	}
	router.GET("/api-docs", apiNote.Handler())
	return apiNote
}

// Use adds middleware to be applied to all subsequent documented routes
func (an *ApiNote) Use(middleware ...gin.HandlerFunc) {
	an.middlewares = append(an.middlewares, middleware...)
}

// DocumentedRoute adds a route with optional schemas, parameters, and middleware
func (an *ApiNote) DocumentedRoute(
	method, path, description string,
	responses map[string]string,
	handler gin.HandlerFunc,
	params []Parameter,
	schemas ...interface{},
) {
	key := method + " " + path
	endpoint := Endpoint{
		Method:      method,
		Path:        path,
		Description: description,
		Responses:   responses,
		Parameters:  params,
		// AuthRequired is true if any middleware is applied at this point
		AuthRequired: len(an.middlewares) > 0,
	}

	if len(schemas) > 0 && schemas[0] != nil {
		endpoint.RequestSchema = schemas[0]
	}
	if len(schemas) > 1 && schemas[1] != nil {
		endpoint.ResponseSchema = schemas[1]
	}

	an.endpoints[key] = endpoint

	// Combine middleware with the handler
	handlers := append(an.middlewares, handler)
	an.router.Handle(method, path, handlers...)
}

// Handler serves the API documentation
func (an *ApiNote) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		html := an.generateHTML()
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, html)
	}
}

// JWTMiddleware validates JWT tokens
func (an *ApiNote) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Expecting "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(an.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", claims["sub"])
		}

		c.Next()
	}
}
