package tracing

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/credentials"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/helpers/color"
	"maphraohom.app/maphraohom-backoffice/pkg/logger"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var AppTracer *MyTracer

func CurrentTracer() *MyTracer {
	return AppTracer
}

type (
	MyTracer struct {
		TraceProvider *sdktrace.TracerProvider
		Tracer        trace.Tracer
		Logger        *logger.Logger
	}
)

func NewTracer(traceProvider *sdktrace.TracerProvider, tracer trace.Tracer) *MyTracer {
	return &MyTracer{
		TraceProvider: traceProvider,
		Tracer:        tracer,
		Logger:        logger.AppLogger,
	}
}

func InitTracer() *MyTracer {
	if config.Global.OpenTelemetry.OtelExporterOTLPEndpoint == "" {
		if !fiber.IsChild() {
			log.Println("[App] OpenTelemetry: Tracing is", color.Format(color.RED, "off!"))
		}

		return NewTracer(nil, nil)
	}

	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if config.Global.OpenTelemetry.OtelInsecureMode {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(config.Global.OpenTelemetry.OtelExporterOTLPEndpoint),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", config.Global.App.ServiceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Println("[App] Tracer could not set resources:", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Check if OpenTelemetry Endpoint is already set
	if config.Global.OpenTelemetry.OtelExporterOTLPEndpoint != "" {
		if !fiber.IsChild() {
			log.Println("[App] OpenTelemetry: Tracing is", color.Format(color.GREEN, "on!"))
		}
	}

	// Set main tracer
	tracer := otel.Tracer(config.Global.App.ServiceName)

	return NewTracer(tp, tracer)
}

func (t MyTracer) TraceStart(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t.TraceProvider == nil {
		return ctx, nil
	}

	return t.Tracer.Start(ctx, spanName, opts...)
}

func (t MyTracer) TraceEnd(span trace.Span) {
	if t.TraceProvider != nil && span != nil {
		span.End()
	}
}

func (t MyTracer) SetAttributes(span trace.Span, attr attribute.KeyValue) {
	if t.TraceProvider != nil && span != nil {
		span.SetAttributes(attr)
	}
}

func (t MyTracer) Cleanup() {
	if t.TraceProvider == nil {
		return
	}

	if err := t.TraceProvider.Shutdown(context.Background()); err != nil {
		log.Printf("[App] Error shutting down tracer provider: %v", err)
	}
}
