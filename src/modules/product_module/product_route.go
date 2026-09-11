package product_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	product := v1.Group("products")
	product.Get("", func(c *fiber.Ctx) error { return m.Controller().GetProducts(c) })
}
