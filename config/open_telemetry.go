package config

import (
	"os"
	"strconv"
)

type openTelemetryConfig struct {
	OtelExporterOTLPEndpoint string
	OtelInsecureMode         bool
}

func NewOpenTelemetryConfig() *openTelemetryConfig {
	otelInsecureMode := func() bool {
		// Default insecure mode is true
		otelInsecureMode := true
		envOtelInsecureMode, err := strconv.ParseBool(os.Getenv("OTEL_INSECURE_MODE"))
		if err == nil {
			otelInsecureMode = envOtelInsecureMode
		}
		return otelInsecureMode
	}()

	return &openTelemetryConfig{
		OtelExporterOTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OtelInsecureMode:         otelInsecureMode,
	}
}
