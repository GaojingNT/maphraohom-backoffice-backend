package platform_type_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	platformType := v1.Group("platform-types", middleware.JwtAuthProtected())
	platformType.Get("", func(c *fiber.Ctx) error { return m.Controller().GetPlatformTypes(c) })
}
