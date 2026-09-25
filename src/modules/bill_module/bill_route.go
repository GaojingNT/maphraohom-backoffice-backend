package bill_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	bill := v1.Group("bills", middleware.JwtAuthProtected())
	bill.Get("", func(c *fiber.Ctx) error { return m.Controller().GetBills(c) })
	bill.Get(":id", func(c *fiber.Ctx) error { return m.Controller().GetBill(c) })
	bill.Post("", func(c *fiber.Ctx) error { return m.Controller().CreateBill(c) })
	bill.Put(":id", func(c *fiber.Ctx) error { return m.Controller().UpdateBill(c) })
	bill.Delete(":id", func(c *fiber.Ctx) error { return m.Controller().DeleteBill(c) })
	bill.Put(":id/slip", func(c *fiber.Ctx) error { return m.Controller().UploadSlip(c) })
	bill.Delete(":id/slip", func(c *fiber.Ctx) error { return m.Controller().DeleteSlip(c) })
}
