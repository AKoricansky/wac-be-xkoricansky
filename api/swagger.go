package api

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

// swaggerTemplate is the HTML template for the Swagger UI
const swaggerTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Ambulance Counseling API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.9.0/favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@5.9.0/favicon-16x16.png" sizes="16x16" />
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    
    *,
    *:before,
    *:after {
      box-sizing: inherit;
    }

    body {
      margin: 0;
      background: #fafafa;
    }
  </style>
</head>

<body>
  <div id="swagger-ui"></div>

  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js" charset="UTF-8"> </script>
  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
  <script>
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        url: "{{.SpecURL}}",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      // End Swagger UI call region

      window.ui = ui;
    };
  </script>
</body>
</html>
`

type SwaggerIndexData struct {
	SpecURL string
}

func HandleSwaggerUI(ctx *gin.Context) {
	tmpl, err := template.New("swagger-ui").Parse(swaggerTemplate)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Failed to parse Swagger UI template")
		return
	}

	data := SwaggerIndexData{
		SpecURL: "/openapi",
	}

	ctx.Status(http.StatusOK)
	ctx.Header("Content-Type", "text/html; charset=utf-8")

	if err := tmpl.Execute(ctx.Writer, data); err != nil {
		ctx.String(http.StatusInternalServerError, "Failed to render Swagger UI template")
		return
	}
}

func RegisterSwaggerRoutes(router *gin.Engine) {
	router.GET("/swagger", HandleSwaggerUI)
}
