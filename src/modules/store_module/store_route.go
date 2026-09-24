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
	store.Put(":id", func(c *fiber.Ctx) error { return m.Controller().UpdateStore(c) })
	store.Put(":id/logo", func(c *fiber.Ctx) error { return m.Controller().UploadLogo(c) })
	store.Delete(":id/logo", func(c *fiber.Ctx) error { return m.Controller().DeleteLogo(c) })
	store.Put(":id/signature", func(c *fiber.Ctx) error { return m.Controller().UploadSignature(c) })
	store.Delete(":id/signature", func(c *fiber.Ctx) error { return m.Controller().DeleteSignature(c) })
	store.Get(":storeId/last-prices", func(c *fiber.Ctx) error { return m.Controller().GetLastPrices(c) })
}
