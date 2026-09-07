package config

import (
	"os"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type fiberConfig struct {
	Config     fiber.Config
	Middleware *fiberMiddlewareConfig
}

type fiberMiddlewareConfig struct {
	ETag    etag.Config
	Cors    cors.Config
	Logger  logger.Config
	Favicon favicon.Config
}

func NewFiberConfig() *fiberConfig {
	var isPrefork bool

	// Note: Disabled for now
	if os.Getenv("FIBER_PREFORK") == "true" {
		isPrefork = true
	} else {
		isPrefork = false
	}

	FiberConfig := fiber.Config{
		Prefork:                 isPrefork,
		CaseSensitive:           true,
		StrictRouting:           true,
		EnableTrustedProxyCheck: true,
		ServerHeader:            "",
		AppName:                 os.Getenv("APP_NAME"),
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
	}

	ETagConfig := etag.Config{
		Weak: true,
	}

	corsAllowCredentials := false
	var corsAllowOrigins string
	if os.Getenv("CORS_ALLOW_ORIGINS") != "" && os.Getenv("CORS_ALLOW_ORIGINS") != "*" {
		corsAllowOrigins = os.Getenv("CORS_ALLOW_ORIGINS")
		corsAllowCredentials = true
	} else {
		corsAllowOrigins = "*"
	}

	CorsConfig := cors.Config{
		Next:             nil,
		AllowOrigins:     corsAllowOrigins,
		AllowCredentials: corsAllowCredentials,
	}

	LoggerConfig := logger.Config{
		Format: "[${ip}]:${port} (${pid}) ${locals:requestid} ${status} - ${method} ${path}\n",
	}

	FaviconConfig := favicon.Config{
		Next: nil,
		File: "",
	}

	return &fiberConfig{
		Config: FiberConfig,
		Middleware: &fiberMiddlewareConfig{
			ETag:    ETagConfig,
			Cors:    CorsConfig,
			Logger:  LoggerConfig,
			Favicon: FaviconConfig,
		},
	}
}
