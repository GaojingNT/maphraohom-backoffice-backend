package file_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	// Wildcard captures the full storage key, e.g. bills/slips/<uuid>.jpeg
	v1.Get("files/*", func(c *fiber.Ctx) error { return m.Controller().GetFile(c) })
}
