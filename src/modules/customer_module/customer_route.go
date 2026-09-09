package customer_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	customer := v1.Group("customers")
	customer.Get("", func(c *fiber.Ctx) error { return m.Controller().GetCustomers(c) })
	customer.Get(":id", func(c *fiber.Ctx) error { return m.Controller().GetCustomer(c) })
	customer.Get(":id/addresses", func(c *fiber.Ctx) error { return m.Controller().GetCustomerAddresses(c) })
	customer.Get(":id/phones", func(c *fiber.Ctx) error { return m.Controller().GetCustomerPhones(c) })
}
