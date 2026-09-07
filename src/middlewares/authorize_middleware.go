package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

func (m *Middleware) HasRoles(roles []string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var (
			_, childSpan = m.tracer.TraceStart(c.Context(), "HasRolesMiddleware", trace.WithAttributes(attribute.String("middleware", "HasRoles")))
			authUser     = c.Locals("authUser").(*models.User)
			err          error
		)

		// Get user data
		if err = m.db.Preload("Role").First(&authUser, authUser.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return exception.HttpErrorResponseMapping(c, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
			} else {
				exception.SqlErrorMessage = err.Error()
				return exception.HttpErrorResponseMapping(c, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
			}
		}

		// Authorize
		isValid := false

		for _, role := range roles {
			if authUser.HasRole(role) {
				isValid = true
				break
			}
		}

		m.tracer.TraceEnd(childSpan)

		if !isValid {
			return exception.HttpErrorResponseMapping(c, fiber.StatusForbidden, exception.ForbiddenResponseError, exception.ErrForbidden)
		}

		return c.Next()
	}
}

func (m *Middleware) HasPermissions(permissions []string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var (
			_, childSpan = m.tracer.TraceStart(c.Context(), "HasPermissionsMiddleware", trace.WithAttributes(attribute.String("middleware", "HasPermissions")))
			authUser     = c.Locals("authUser").(*models.User)
			err          error
		)

		// Get user data
		if err = m.db.Preload("Role").Preload("Role.Permissions").First(&authUser, authUser.ID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return exception.HttpErrorResponseMapping(c, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
			} else {
				exception.SqlErrorMessage = err.Error()
				return exception.HttpErrorResponseMapping(c, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
			}
		}

		// Authorize
		isValid := false

		for _, permission := range permissions {
			if authUser.HasPermission(permission) {
				isValid = true
				break
			}
		}

		m.tracer.TraceEnd(childSpan)

		if !isValid {
			return exception.HttpErrorResponseMapping(c, fiber.StatusForbidden, exception.ForbiddenResponseError, exception.ErrForbidden)
		}

		return c.Next()
	}
}
