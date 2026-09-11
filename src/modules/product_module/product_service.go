package product_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/src/modules/product_module/responses"
)

func (s Service) GetProducts(ctx context.Context) ([]responses.ProductListItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetProductsService", trace.WithAttributes(attribute.String("service", "GetProducts")))

	products, err := s.productRepository().GetProducts(ctx)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return responses.ProductListItem{}.Collection(products), nil
}
