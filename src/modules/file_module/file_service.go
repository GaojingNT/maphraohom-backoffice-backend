package file_module

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s Service) GetFile(ctx context.Context, key string) (*os.File, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetFileService", trace.WithAttributes(attribute.String("service", "GetFile"), attribute.String("key", key)))

	file, err := s.fileSystem.Get(ctx, key)

	s.tracer.TraceEnd(childSpan)

	return file, err
}
