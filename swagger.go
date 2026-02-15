package notelink

import (
	"github.com/gofiber/fiber/v3"
)

// SwaggerUIHandler returns a handler that serves the Swagger UI
// The Swagger UI is loaded from CDN and points to /api-docs/openapi.json
func (an *ApiNote) SwaggerUIHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		html := an.generateSwaggerHTML()
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}

// ScalarUIHandler returns a handler that serves the Scalar API documentation UI
// Scalar is a modern alternative to Swagger UI with a cleaner interface
func (an *ApiNote) ScalarUIHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		html := an.generateScalarHTML()
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}

// generateSwaggerHTML creates the Swagger UI HTML page
func (an *ApiNote) generateSwaggerHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + an.config.Title + ` - Swagger UI</title>
    <link rel="icon" type="image/png" sizes="32x32" href="/icon.png">
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
        .topbar {
            display: none;
        }
        .swagger-ui .info {
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>

    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/api-docs/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                persistAuthorization: true,
                tryItOutEnabled: true
            });

            window.ui = ui;
        };
    </script>
</body>
</html>`
}

// generateScalarHTML creates the Scalar UI HTML page
func (an *ApiNote) generateScalarHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + an.config.Title + ` - Scalar API Documentation</title>
    <link rel="icon" type="image/png" sizes="32x32" href="/icon.png">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <script
        id="api-reference"
        data-url="/api-docs/openapi.json"
        data-configuration='{"theme":"deepSpace","showSidebar":true,"hideDarkModeToggle":false,"hideModels":false,"hideDownloadButton":false,"searchHotKey":"k"}'
    ></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`
}
