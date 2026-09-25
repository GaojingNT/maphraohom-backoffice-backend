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

					// Sessions last up to AUTH_SESSION_LIFETIME (30 days by
					// default), so a valid signature isn't enough — the user
					// must still exist (soft-deleted users are excluded by
					// GORM's default scope), or deleting a user wouldn't lock
					// them out until their token expired.
					var exists int64
					if err := m.db.Model(&models.User{}).Where("id = ?", user.ID).Count(&exists).Error; err != nil {
						exception.SqlErrorMessage = err.Error()
						return exception.HttpErrorResponseMapping(c, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
					}
					if exists == 0 {
						return exception.HttpErrorResponseMapping(c, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
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
