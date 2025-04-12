# Notelink

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)

Notelink is a Go package that simplifies API documentation generation and integrates it with a Fiber web server. It allows developers to define API endpoints with parameters, schemas, and authentication, and automatically generates an interactive HTML documentation page.

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fcanvas-tech-horizon%2Fnotelink.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fcanvas-tech-horizon%2Fnotelink?ref=badge_large&issueType=license)

## Features

- **Endpoint Documentation**: Define HTTP methods, paths, descriptions, and responses.
- **Parameter Support**: Specify query, path, and header parameters.
- **Schema Generation**: Automatically generate TypeScript interfaces from Go structs.
- **JWT Authentication**: Built-in middleware for token validation.
- **Interactive Testing**: Test endpoints directly from the generated HTML docs.

![image](https://github.com/user-attachments/assets/46c5857c-18d8-4ce2-a62c-28ad1e508ffc)

## Installation

To use Notelink, ensure you have Go installed, then add it to your project:

```bash
go get github.com/canvas-tech-horizon/notelink
```
## Dependencies
Notelink requires the following dependencies:
```go
require (
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/golang-jwt/jwt/v5 v5.2.1
    github.com/joho/godotenv v1.5.1 // optional, for .env support
)
```
Run `go mod tidy` to install them.

## Usage

Here's a complete example to get you started:

Example: Combined API
Create a `main.go` file:
```go
package main

import (
	"log"
	"os"

	"github.com/canvas-tech-horizon/notelink"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "my-secret-key"
	}

	config := notelink.Config{
		Title:       "Sample API",
		Description: "A sample API with documentation",
		Version:     "1.0.0",
		Host:        "localhost:8080",
		BasePath:    "/api",
	}

	api := notelink.NewApiNote(&config, jwtSecret)

	api.DocumentedRoute(
		"GET",
		"/v1/users",
		"List users",
		map[string]string{
			"200": "Success",
			"401": "Unauthorized",
		},
		func(c *fiber.Ctx) error {
			users := []UserResponse{
				{ID: 1, Name: "John", Email: "john@example.com"},
			}
			return c.JSON(users)
		},
		[]notelink.Parameter{
			{Name: "limit", In: "query", Type: "number", Description: "Max users", Required: false},
		},
		nil,
		[]UserResponse{},
	)

	if err := api.Listen(); err != nil {
		panic(err)
	}
}
```

## Configuration
Create a `.env` file for sensitive data:
```text
JWT_SECRET=your-secure-secret-key
```
## Running the Application
```go
go run main.go
```
Visit `http://localhost:8080/api-docs` to see the interactive documentation.

## API Documentation
The package generates an HTML page with:

Collapsible Sections: Organized by API version and endpoint paths.
- Method Coloring: GET (green), POST (blue), etc.
- Parameter Details: Lists all parameters with types and descriptions.
- Schemas: Displays TypeScript interfaces for request/response bodies.
- API Testing: Forms to test endpoints directly from the browser.

## License
This project is licensed under the MIT License - see the  file for details.
