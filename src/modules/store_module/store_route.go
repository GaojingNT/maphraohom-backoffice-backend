package store_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	store := v1.Group("stores")
	store.Get("", func(c *fiber.Ctx) error { return m.Controller().GetStores(c) })
	store.Get(":id", func(c *fiber.Ctx) error { return m.Controller().GetStore(c) })
	store.Get(":id/products", func(c *fiber.Ctx) error { return m.Controller().GetStoreProducts(c) })
	store.Get(":id/products/:productId", func(c *fiber.Ctx) error { return m.Controller().GetStoreProduct(c) })
	store.Put(":id/products/:productId", func(c *fiber.Ctx) error { return m.Controller().UpdateStoreProductPrice(c) })
	store.Get(":id/base-prices", func(c *fiber.Ctx) error { return m.Controller().GetStoreBasePrices(c) })
}
