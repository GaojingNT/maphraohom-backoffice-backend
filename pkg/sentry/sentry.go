package exceptions

import (
	"log"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/helpers/color"
)

func Initialize() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.Global.Sentry.SentryDSN,
		EnableTracing:    config.Global.Sentry.SentryEnableTracing,
		TracesSampleRate: config.Global.Sentry.SentryTracesSampleRate,
	})

	// If there is an error, do not continue.
	if err != nil {
		log.Fatalf("[App] sentry.Init: %s", err)
	}

	// Check if Sentry DSN is already set
	if config.Global.Sentry.SentryDSN != "" {
		if !fiber.IsChild() {
			log.Println("[App] Sentry: Error handling is", color.Format(color.GREEN, "on!"))
			if config.Global.Sentry.SentryEnableTracing {
				log.Println("[App] Sentry: Tracing is", color.Format(color.GREEN, "on!"))
				log.Println("[App] Sentry: Tracing sample rate is", config.Global.Sentry.SentryTracesSampleRate)
			}
		}
	}
}
