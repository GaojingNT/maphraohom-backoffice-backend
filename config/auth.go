package config

import (
	"os"
	"strconv"
)

type authConfig struct {
	KeycloakConfig  *keycloakConfig
	SecretKey       string
	SessionLifetime int
	PublicKeyPath   string
	PrivateKeyPath  string
}

type keycloakConfig struct {
	Endpoint string
	Realm    string
}

func NewAuthConfig() *authConfig {
	return &authConfig{
		KeycloakConfig: NewKeycloakConfig(),
		SecretKey:      os.Getenv("AUTH_SECRET_KEY"),
		SessionLifetime: func() int {
			// default 3600 seconds
			sessionLifeTime := 3600
			envSessionLifeTime, err := strconv.Atoi(os.Getenv("AUTH_SESSION_LIFETIME"))
			if err == nil {
				sessionLifeTime = envSessionLifeTime
			}
			return sessionLifeTime
		}(),
	}
}

func NewKeycloakConfig() *keycloakConfig {
	return &keycloakConfig{
		Endpoint: os.Getenv("AUTH_KEYCLOAK_ENDPOINT"),
		Realm:    os.Getenv("AUTH_KEYCLOAK_REALM"),
	}
}
