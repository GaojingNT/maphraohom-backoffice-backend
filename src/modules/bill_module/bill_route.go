package bill_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	bill := v1.Group("bills")
	bill.Get("", func(c *fiber.Ctx) error { return m.Controller().GetBills(c) })
	bill.Get(":id", func(c *fiber.Ctx) error { return m.Controller().GetBill(c) })
}
