package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

func (m *Middleware) KeycloakProtected() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Note: Check for SSO access token is exist
		if c.Get("Authorization") == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		} else if c.Get("authorization") == "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}

		// TODO: Validate the SSO access token with public key

		return c.Next()
	}
}
