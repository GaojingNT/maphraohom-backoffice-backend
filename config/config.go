package config

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	Global       *Config
	IsProduction bool
)

type Config struct {
	App           *appConfig
	Auth          *authConfig
	Fiber         *fiberConfig
	Database      *databaseConfig
	Redis         *redisConfig
	Sentry        *sentryConfig
	OpenTelemetry *openTelemetryConfig
	Logger        *loggerConfig
	FileSystem    *fileSystemConfig
	Mail          *mailConfig
}

func NewConfig() *Config {
	// If ENV variable is not set, default is "local"
	// local = load from .env.dev file
	// production = load from .env file
	if os.Getenv("ENV") == "" || os.Getenv("ENV") == "local" || os.Getenv("ENV") == "development" {
		// Load .env.dev file
		godotenv.Load(".env.dev")
	} else {
		// Production environment
		// Load .env file
		godotenv.Load(".env")
		IsProduction = true
	}

	return &Config{
		App:           NewAppConfig(),
		Auth:          NewAuthConfig(),
		Fiber:         NewFiberConfig(),
		Database:      NewDatabaseConfig(),
		Redis:         NewRedisConfig(),
		Sentry:        NewSentryConfig(),
		OpenTelemetry: NewOpenTelemetryConfig(),
		Logger:        NewLoggerConfig(),
		FileSystem:    NewFileSystemConfig(),
		Mail:          NewMailConfig(),
	}
}
