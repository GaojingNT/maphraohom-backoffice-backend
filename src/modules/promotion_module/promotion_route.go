package promotion_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	promotion := v1.Group("promotions")
	promotion.Get("", func(c *fiber.Ctx) error { return m.Controller().GetPromotions(c) })
	promotion.Post("", func(c *fiber.Ctx) error { return m.Controller().CreatePromotion(c) })
}
