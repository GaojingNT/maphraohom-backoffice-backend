package middlewares

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

func (m *Middleware) JwtAuthProtected() func(*fiber.Ctx) error {
	return jwtware.New(jwtware.Config{
		// Set the JWT signing key
		SigningKey: jwtware.SigningKey{Key: []byte(config.Global.Auth.SecretKey)},

		// Set lookup function
		TokenLookup: "header:Authorization",

		// Set auth scheme
		AuthScheme: "Bearer",

		// Set context key
		ContextKey: "auth",

		// Handle success case
		SuccessHandler: func(c *fiber.Ctx) error {
			// Get user from context
			token := c.Locals("auth").(*jwt.Token)

			// using claims data for get user from database
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				return exception.HttpErrorResponseMapping(c, fiber.StatusBadRequest, exception.InvalidTokenClaimResponseError, exception.ErrInvalidTokenClaim)
			}

			// Mapping content
			if content, ok := claims["dat"].(map[string]interface{}); ok {
				// Get user model
				if contentUser, ok := content["user"].(map[string]interface{}); ok {
					user := new(models.User)

					// Decode to struct
					user.ID = int(contentUser["id"].(float64))
					user.Email = contentUser["email"].(string)

					// Role
					if role, ok := contentUser["role"].(map[string]interface{}); ok {
						user.Role = new(models.Role)
						user.Role.ID = int(role["id"].(float64))
						user.Role.Name = role["name"].(string)
					}

					// Set auth user to context
					c.Locals("authUser", user)
				} else {
					return exception.HttpErrorResponseMapping(c, fiber.StatusBadRequest, exception.InvalidUserDataResponseError, exception.ErrInvalidUserData)
				}
			} else {
				return exception.HttpErrorResponseMapping(c, fiber.StatusBadRequest, exception.InvalidTokenClaimDataResponseError, exception.ErrInvalidTokenClaimData)
			}

			return c.Next()
		},

		// Handle error cases
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return exception.HttpErrorResponseMapping(c, fiber.StatusUnauthorized, exception.InvalidTokenResponseError, exception.ErrUnauthorized)
		},
	})
}
