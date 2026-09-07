package config

import (
	"os"
	"strconv"
)

type mailConfig struct {
	SmtpHost     string
	SmtpPort     int
	SmtpUsername string
	SmtpPassword string
}

func NewMailConfig() *mailConfig {
	return &mailConfig{
		SmtpHost: os.Getenv("MAIL_SMTP_HOST"),
		SmtpPort: func() int {
			port := 587
			envSmtpPort, err := strconv.Atoi(os.Getenv("MAIL_SMTP_PORT"))
			if err == nil {
				port = envSmtpPort
			}
			return port
		}(),
		SmtpUsername: os.Getenv("MAIL_SMTP_USERNAME"),
		SmtpPassword: os.Getenv("MAIL_SMTP_PASSWORD"),
	}
}
