package {{.ModuleName}}_module

import (
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/src/middlewares"
)

// Register controller routes
func (m module) Routes(route fiber.Router, middleware middlewares.IMiddleware) {
	// Version 1
	v1 := route.Group("v1")
	{{.ModuleName}} := v1.Group("{{.ModulePluralName}}")
	{{.ModuleName}}.Get("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"{{.ModuleName}}:view"}), func(c *fiber.Ctx) error { return m.Controller().Get{{.PascalModulePluralName}}(c) })
	{{.ModuleName}}.Get(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"{{.ModuleName}}:view"}), func(c *fiber.Ctx) error { return m.Controller().Get{{.PascalModuleName}}(c) })
	{{.ModuleName}}.Post("", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"{{.ModuleName}}:create"}), func(c *fiber.Ctx) error { return m.Controller().Create{{.PascalModuleName}}(c) })
	{{.ModuleName}}.Put(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"{{.ModuleName}}:update"}), func(c *fiber.Ctx) error { return m.Controller().Update{{.PascalModuleName}}(c) })
	{{.ModuleName}}.Delete(":id", middleware.JwtAuthProtected(), middleware.HasPermissions([]string{"{{.ModuleName}}:destroy"}), func(c *fiber.Ctx) error { return m.Controller().Delete{{.PascalModuleName}}(c) })
}
