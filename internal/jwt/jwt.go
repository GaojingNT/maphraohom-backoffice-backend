package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

var JwtToken JWT

type JWT struct {
	secretKey []byte
}

func NewJWT() JWT {
	return JWT{
		secretKey: []byte(config.Global.Auth.SecretKey),
	}
}

func (j JWT) Create(ttl time.Duration, content interface{}) (string, error) {
	now := time.Now()

	claims := make(jwt.MapClaims)
	claims["dat"] = content             // Our custom data.
	claims["exp"] = now.Add(ttl).Unix() // The expiration time after which the token must be disregarded.
	claims["iat"] = now.Unix()          // The time at which the token was issued.
	claims["nbf"] = now.Unix()          // The time before which the token must be disregarded.

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("NewWithClaimsError: %w", err)
	}

	return token, nil
}

func (j JWT) Payload(accessToken string) (map[string]interface{}, error) {
	var content map[string]interface{}

	// Verify the signature
	token, err := jwt.Parse(accessToken, func(jwtToken *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		return nil, exception.ErrInvalidToken
	}

	// using claims data for get user from database
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, exception.ErrInvalidTokenClaim
	}

	// Check token expire
	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	currentTime := time.Now()

	if expirationTime.Before(currentTime) {
		return nil, exception.ErrTokenExpired
	}

	// Mapping content
	if content, ok = claims["dat"].(map[string]interface{}); !ok {
		return nil, exception.ErrInvalidTokenClaim
	}

	return content, nil
}
