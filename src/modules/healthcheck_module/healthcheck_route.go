package healthcheck_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	route.Get("/health", func(c *fiber.Ctx) error { return m.Controller().CheckDatabaseConnection(c) })
}
