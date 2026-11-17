package platform_type_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/src/modules/platform_type_module/responses"
)

func (s Service) GetPlatformTypes(ctx context.Context) ([]responses.PlatformTypeResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPlatformTypesService", trace.WithAttributes(attribute.String("service", "GetPlatformTypes")))

	platformTypes, err := s.platformTypeRepository().GetActivePlatformTypes(ctx)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return responses.Collection(platformTypes), nil
}
