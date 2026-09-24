package store_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/store_module/responses"
)

func (s Service) GetStores(ctx context.Context, paginate *paginator.Pagination) (*paginator.Pagination, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetStoresService", trace.WithAttributes(attribute.String("service", "GetStores")))

	result, err := s.storeRepository().GetStorePaginate(ctx, paginate)

	if result != nil && result.Data != nil {
		result.Data = responses.StoreListItem{}.Collection(result.Data.([]models.Store))
	}

	s.tracer.TraceEnd(childSpan)

	return result, err
}

func (s Service) GetStore(ctx context.Context, id int) (*responses.StoreDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetStoreService", trace.WithAttributes(attribute.String("service", "GetStore")))

	store, err := s.storeRepository().GetStoreByID(ctx, id)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.StoreDetailResponse).Make(store), nil
}

// GetLastPrices returns, for every product, the price used in this store's
// most recent bill of the given type — used to prefill the create-bill form.
func (s Service) GetLastPrices(ctx context.Context, storeID int, billType string) ([]responses.LastPriceItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetLastPricesService", trace.WithAttributes(attribute.String("service", "GetLastPrices")))

	rows, err := s.storeRepository().GetLastPrices(ctx, storeID, billType)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	items := make([]responses.LastPriceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, responses.LastPriceItem{
			ProductID: row.ProductID,
			Price:     row.Price,
		})
	}

	return items, nil
}
