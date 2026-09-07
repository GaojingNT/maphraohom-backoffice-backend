package config

import (
	"os"
	"strconv"
)

type sentryConfig struct {
	SentryDSN              string
	SentryEnableTracing    bool
	SentryTracesSampleRate float64
}

func NewSentryConfig() *sentryConfig {
	sentryEnableTracing := func() bool {
		// Default is false
		sentryEnableTracing := false
		envSentryEnableTracing, err := strconv.ParseBool(os.Getenv("SENTRY_ENABLE_TRACING"))
		if err == nil {
			sentryEnableTracing = envSentryEnableTracing
		}
		return sentryEnableTracing
	}()

	sentryTracesSampleRate := func() float64 {
		// Default traces sample rate is 0.2
		sentryTracesSampleRate := 0.2
		envSentryTracesSampleRate, err := strconv.ParseFloat(os.Getenv("SENTRY_TRACES_SAMPLE_RATE"), 64)
		if err == nil {
			sentryTracesSampleRate = envSentryTracesSampleRate
		}
		return sentryTracesSampleRate
	}()

	return &sentryConfig{
		SentryDSN:              os.Getenv("SENTRY_DSN"),
		SentryEnableTracing:    sentryEnableTracing,
		SentryTracesSampleRate: sentryTracesSampleRate,
	}
}
