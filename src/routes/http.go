package routes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	fiberCors "github.com/gofiber/fiber/v2/middleware/cors"
	fiberEtag "github.com/gofiber/fiber/v2/middleware/etag"
	fiberFavicon "github.com/gofiber/fiber/v2/middleware/favicon"
	fiberLoggerMiddleware "github.com/gofiber/fiber/v2/middleware/logger"
	fiberMonitor "github.com/gofiber/fiber/v2/middleware/monitor"
	fiberPprof "github.com/gofiber/fiber/v2/middleware/pprof"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
	fiberRequestID "github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"maphraohom.app/maphraohom-backoffice/app/http_server"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
	"maphraohom.app/maphraohom-backoffice/src/modules/auth_module"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module"
	"maphraohom.app/maphraohom-backoffice/src/modules/customer_module"
	"maphraohom.app/maphraohom-backoffice/src/modules/file_module"
	"maphraohom.app/maphraohom-backoffice/src/modules/healthcheck_module"
	"maphraohom.app/maphraohom-backoffice/src/modules/store_module"
	"maphraohom.app/maphraohom-backoffice/src/modules/user_module"

	// swagger handler
	_ "maphraohom.app/maphraohom-backoffice/docs"
)

func HTTPRootMiddleware(r *http_server.Route) {
	// Default middleware configs
	r.Use(
		fiberRequestID.New(),
		fiberEtag.New(config.Global.Fiber.Middleware.ETag),
		fiberCors.New(config.Global.Fiber.Middleware.Cors),
		fiberLoggerMiddleware.New(config.Global.Fiber.Middleware.Logger),
		fiberFavicon.New(config.Global.Fiber.Middleware.Favicon),
		fiberRecover.New(),
		fiberPprof.New(), // pprof is registered at "/debug/pprof"
	)
}

func HTTPRoutes(s *http_server.HttpServer) {
	// Middlewares
	HTTPRootMiddleware(s.Router)
	middleware := middlewares.NewMiddleware()

	// REST API endpoints ------------------------------------------------------------------

	s.Route().Get("/", func(c *fiber.Ctx) error { return c.Status(fiber.StatusOK).SendString(c.App().Config().AppName) })
	s.Route().Get("/monitor", fiberMonitor.New(fiberMonitor.Config{Title: "App Monitoring"}))
	s.Route().Get("/swagger/*", swagger.HandlerDefault)

	// Cache clear endpoint
	// This endpoint is used to clear the Redis cache, useful for development or debugging purposes.
	s.Route().Get("/cache/clear", func(c *fiber.Ctx) error {
		if err := s.Cacher.Redis.FlushAll(c.Context()); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("ClearCacheFailedError: %+v\n", err))
		}
		return c.SendStatus(fiber.StatusOK)
	})

	// API Grouping
	api := s.Route().Group("api")

	// Register routes
	healthcheck_module.NewModule().Routes(s.MainRoute(), middleware)
	auth_module.NewModule().Routes(api, middleware)
	user_module.NewModule().Routes(api, middleware)
	bill_module.NewModule().Routes(api, middleware)
	store_module.NewModule().Routes(api, middleware)
	customer_module.NewModule().Routes(api, middleware)
	file_module.NewModule().Routes(api, middleware)
}
