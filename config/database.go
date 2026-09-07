package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type databaseConfig struct {
	DatabaseDriver string
	DatabaseDSN    string
	DatabaseURL    string
	// Additional
	DatabaseMaxIdleConns int
	DatabaseMaxOpenConns int
}

func NewDatabaseConfig() *databaseConfig {
	databaseDriver := func() string {
		driver := os.Getenv("DATABASE_DRIVER")
		if driver != "" {
			return driver
		}

		// Default database driver is PostgreSQL
		return "postgresql"
	}()

	databaseDSN := func() string {
		// Encoding password for postgresql URI schema
		password := os.Getenv("DATABASE_PASSWORD")
		encodedPassword := url.QueryEscape(password)

		switch databaseDriver {
		case "mysql":
			return fmt.Sprintf(
				`%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local`,
				os.Getenv("DATABASE_USER"),
				encodedPassword,
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
			)
		case "sqlserver":
			return fmt.Sprintf(
				`sqlserver://%s:%s@%s:%s?database=%s&trustServerCertificate=true`,
				os.Getenv("DATABASE_USER"),
				encodedPassword,
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
			)
		default:
			return fmt.Sprintf(
				`host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s`,
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_USER"),
				os.Getenv("DATABASE_PASSWORD"),
				os.Getenv("DATABASE_NAME"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_SSL_MODE"),
				os.Getenv("DATABASE_TIMEZONE"),
			)
		}
	}()

	databaseURL := func() string {
		var sslMode string
		// Check SSL mode
		sslMode = os.Getenv("DATABASE_SSL_MODE")
		if sslMode == "" || sslMode == "disable" {
			sslMode = "disable"
		} else if sslMode == "prefer" {
			// Set SSL mode to require (postgresql schema 'sslmode' does not support prefer mode, using require instead)
			sslMode = "require"
		}

		// Encoding password for postgresql URI schema
		password := os.Getenv("DATABASE_PASSWORD")
		encodedPassword := url.QueryEscape(password)

		switch databaseDriver {
		case "mysql":
			return fmt.Sprintf(
				`%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local`,
				os.Getenv("DATABASE_USER"),
				encodedPassword,
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
			)
		case "sqlserver":
			return fmt.Sprintf(
				`sqlserver://%s:%s@%s:%s?database=%s`,
				os.Getenv("DATABASE_USER"),
				encodedPassword,
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
			)
		default:
			return fmt.Sprintf(
				`postgres://%s:%s@%s:%s/%s?sslmode=%s`,
				os.Getenv("DATABASE_USER"),
				encodedPassword,
				os.Getenv("DATABASE_HOST"),
				os.Getenv("DATABASE_PORT"),
				os.Getenv("DATABASE_NAME"),
				sslMode,
			)
		}
	}()

	databaseMaxIdleConns := func() int {
		// Default max idle conns is 2
		databaseMaxIdleConns := 2
		envDatabaseMaxIdleConns, err := strconv.Atoi(os.Getenv("DATABASE_MAX_IDLE_CONNS"))
		if err == nil {
			databaseMaxIdleConns = envDatabaseMaxIdleConns
		}
		return databaseMaxIdleConns
	}()

	databaseMaxOpenConns := func() int {
		// Default max open conns is 3
		databaseMaxOpenConns := 3
		envDatabaseMaxOpenConns, err := strconv.Atoi(os.Getenv("DATABASE_MAX_OPEN_CONNS"))
		if err == nil {
			databaseMaxOpenConns = envDatabaseMaxOpenConns
		}
		return databaseMaxOpenConns
	}()

	return &databaseConfig{
		DatabaseDriver:       databaseDriver,
		DatabaseDSN:          databaseDSN,
		DatabaseURL:          databaseURL,
		DatabaseMaxIdleConns: databaseMaxIdleConns,
		DatabaseMaxOpenConns: databaseMaxOpenConns,
	}
}
