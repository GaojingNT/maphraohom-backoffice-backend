package auth_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware *middlewares.Middleware) {
	// Version 1
	v1 := route.Group("v1")
	auth := v1.Group("auth")
	auth.Post("/sign-in", func(c *fiber.Ctx) error { return m.Controller().SignIn(c) })
	auth.Post("/forgot-password", func(c *fiber.Ctx) error { return m.Controller().ForgotPassword(c) })
	auth.Post("/reset-password", func(c *fiber.Ctx) error { return m.Controller().ResetPassword(c) })
	auth.Post("/encrypt-password", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().EncryptPassword(c) })
	auth.Get("/profile", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().GetProfile(c) })
	auth.Put("/profile", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().UpdateProfile(c) })
	auth.Put("/profile/password", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().ChangePassword(c) })
	auth.Put("/profile/signature", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().UploadSignature(c) })
	auth.Delete("/profile/signature", middleware.JwtAuthProtected(), func(c *fiber.Ctx) error { return m.Controller().DeleteSignature(c) })
}
