package notelink

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// Benchmark structures
type BenchUser struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Age      int     `json:"age"`
	IsActive bool    `json:"is_active"`
	Salary   float64 `json:"salary"`
}

type BenchOrder struct {
	ID       int         `json:"id"`
	UserID   int         `json:"user_id"`
	Total    float64     `json:"total"`
	Items    []BenchItem `json:"items"`
	Customer BenchUser   `json:"customer"`
}

type BenchItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// BenchmarkValidateParametersSmall benchmarks small parameter validation
func BenchmarkValidateParametersSmall(b *testing.B) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		params := []Parameter{
			{Name: "id", In: "query", Type: "integer", Required: true},
			{Name: "name", In: "query", Type: "string", Required: false},
		}
		err := ValidateParameters(c, params)
		if err != nil {
			return c.Status(400).JSON(err)
		}
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test?id=123&name=test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkValidateParametersLarge benchmarks large parameter validation
func BenchmarkValidateParametersLarge(b *testing.B) {
	app := fiber.New()
	params := []Parameter{
		{Name: "id", In: "query", Type: "integer", Required: true},
		{Name: "name", In: "query", Type: "string", Required: true},
		{Name: "email", In: "query", Type: "string", Required: true},
		{Name: "age", In: "query", Type: "integer", Required: false},
		{Name: "active", In: "query", Type: "boolean", Required: false},
		{Name: "score", In: "query", Type: "number", Required: false},
		{Name: "token", In: "header", Type: "string", Required: false},
		{Name: "api-key", In: "header", Type: "string", Required: false},
	}

	app.Get("/test", func(c fiber.Ctx) error {
		err := ValidateParameters(c, params)
		if err != nil {
			return c.Status(400).JSON(err)
		}
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test?id=123&name=john&email=john@example.com&age=25&active=true&score=95.5", nil)
	req.Header.Set("token", "abc123")
	req.Header.Set("api-key", "key456")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkValidateRequestBodySimple benchmarks simple struct validation
func BenchmarkValidateRequestBodySimple(b *testing.B) {
	app := fiber.New()
	body := `{"id":1,"name":"John","email":"john@example.com","age":25,"is_active":true,"salary":50000.50}`

	app.Post("/test", func(c fiber.Ctx) error {
		err := ValidateRequestBody(c, BenchUser{})
		if err != nil {
			return c.Status(400).JSON(err)
		}
		return c.SendString("OK")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkValidateRequestBodyNested benchmarks nested struct validation
func BenchmarkValidateRequestBodyNested(b *testing.B) {
	app := fiber.New()
	body := `{
		"id": 1,
		"user_id": 123,
		"total": 299.97,
		"items": [
			{"id": 1, "name": "Item 1", "price": 99.99, "quantity": 1},
			{"id": 2, "name": "Item 2", "price": 199.98, "quantity": 2}
		],
		"customer": {
			"id": 123,
			"name": "John Doe",
			"email": "john@example.com",
			"age": 30,
			"is_active": true,
			"salary": 75000.00
		}
	}`

	app.Post("/test", func(c fiber.Ctx) error {
		err := ValidateRequestBody(c, BenchOrder{})
		if err != nil {
			return c.Status(400).JSON(err)
		}
		return c.SendString("OK")
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}

// BenchmarkGenerateTypeScriptSchemaSimple benchmarks simple TypeScript generation
func BenchmarkGenerateTypeScriptSchemaSimple(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateTypeScriptSchema("User", BenchUser{})
	}
}

// BenchmarkGenerateTypeScriptSchemaNested benchmarks nested TypeScript generation
func BenchmarkGenerateTypeScriptSchemaNested(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateTypeScriptSchema("Order", BenchOrder{})
	}
}

// BenchmarkGenerateJSONTemplateSimple benchmarks simple JSON template generation
func BenchmarkGenerateJSONTemplateSimple(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generateJSONTemplate(BenchUser{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateJSONTemplateNested benchmarks nested JSON template generation
func BenchmarkGenerateJSONTemplateNested(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generateJSONTemplate(BenchOrder{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEscapeHTML benchmarks HTML escaping
func BenchmarkEscapeHTML(b *testing.B) {
	input := `<script>alert("XSS");</script> & "test" 'value'`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = escapeHTML(input)
	}
}

// BenchmarkEscapeJavaScript benchmarks JavaScript escaping
func BenchmarkEscapeJavaScript(b *testing.B) {
	input := `var x = "test"; alert('XSS'); \n\r\t <>&`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = escapeJavaScript(input)
	}
}

// BenchmarkHTMLGenerationSmall benchmarks HTML generation with small endpoint list
func BenchmarkHTMLGenerationSmall(b *testing.B) {
	config := &Config{
		Title:       "Test API",
		Description: "A test API for benchmarking",
		Version:     "1.0.0",
		Host:        "localhost:8080",
	}

	api := &ApiNote{
		config: config,
		endpoints: map[string]Endpoint{
			"GET/users":     {Method: "GET", Path: "/users", Description: "Get all users"},
			"POST/users":    {Method: "POST", Path: "/users", Description: "Create a user"},
			"GET/users/:id": {Method: "GET", Path: "/users/:id", Description: "Get a user by ID"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = api.generateHTML()
	}
}

// BenchmarkHTMLGenerationLarge benchmarks HTML generation with large endpoint list
func BenchmarkHTMLGenerationLarge(b *testing.B) {
	config := &Config{
		Title:       "Test API",
		Description: "A test API for benchmarking",
		Version:     "1.0.0",
		Host:        "localhost:8080",
	}

	endpoints := make(map[string]Endpoint)
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	resources := []string{"users", "orders", "products", "categories", "reviews"}

	for _, resource := range resources {
		for _, method := range methods {
			key1 := method + "/" + resource
			endpoints[key1] = Endpoint{
				Method:      method,
				Path:        "/" + resource,
				Description: method + " " + resource,
				Parameters: []Parameter{
					{Name: "filter", In: "query", Type: "string", Required: false},
				},
			}

			key2 := method + "/" + resource + "/:id"
			endpoints[key2] = Endpoint{
				Method:      method,
				Path:        "/" + resource + "/:id",
				Description: method + " " + resource + " by ID",
				Parameters: []Parameter{
					{Name: "id", In: "path", Type: "integer", Required: true},
				},
			}
		}
	}

	api := &ApiNote{
		config:    config,
		endpoints: endpoints,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = api.generateHTML()
	}
}
