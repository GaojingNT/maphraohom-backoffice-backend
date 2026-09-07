package user_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	user := v1.Group("users")
	user.Get("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"user:view"}), func(c *fiber.Ctx) error { return m.Controller().GetUsers(c) })
	user.Get(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"user:view"}), func(c *fiber.Ctx) error { return m.Controller().GetUser(c) })
	user.Post("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"user:create"}), func(c *fiber.Ctx) error { return m.Controller().CreateUser(c) })
	user.Put(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"user:update"}), func(c *fiber.Ctx) error { return m.Controller().UpdateUser(c) })
	user.Delete(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"user:destroy"}), func(c *fiber.Ctx) error { return m.Controller().DeleteUser(c) })
}
