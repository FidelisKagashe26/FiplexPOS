package server

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func SetupFrontend(app *App) {
	// "./frontend/dist" is where the Docker image places the build;
	// "../frontend/dist" is used when running locally from the backend/ directory.
	distPath := ""
	for _, candidate := range []string{"./frontend/dist", "../frontend/dist"} {
		if _, err := os.Stat(candidate); err == nil {
			distPath = candidate
			break
		}
	}

	if distPath == "" {
		app.Logger.Warn("Frontend dist directory not found, skipping frontend setup. Run 'cd frontend && npm run build' to generate it.")
		return
	}

	app.FiberApp.Use("/", static.New(distPath, static.Config{
		Compress: true,
		Browse:   false,
	}))

	app.FiberApp.Get("/*", func(c fiber.Ctx) error {
		path := c.Path()

		if strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/swagger/") ||
			path == "/healthz" {
			return c.Next()
		}

		return c.SendFile(distPath + "/index.html")
	})
}
