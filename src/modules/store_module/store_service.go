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

func (s Service) GetStoreProducts(ctx context.Context, storeID int) ([]responses.StoreProductPriceItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetStoreProductsService", trace.WithAttributes(attribute.String("service", "GetStoreProducts")))

	prices, err := s.storeRepository().GetStoreProducts(ctx, storeID)
	if err != nil {
		s.tracer.TraceEnd(childSpan)
		return nil, err
	}

	promotion, err := s.storeRepository().GetActivePromotion(ctx, storeID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return responses.StoreProductPriceItem{}.Collection(prices, promotion), nil
}

func (s Service) GetStoreProduct(ctx context.Context, storeID int, productID int) (*responses.StoreProductPriceItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetStoreProductService", trace.WithAttributes(attribute.String("service", "GetStoreProduct")))

	product, resolved, err := s.storeRepository().GetStoreProduct(ctx, storeID, productID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	item := responses.StoreProductPriceItem{
		ProductID:   product.ID,
		ProductName: product.Name,
		Unit:        product.Unit,
		Price:       resolved.Price,
		IsPromotion: resolved.IsPromotion,
		PromotionID: resolved.PromotionID,
	}
	return &item, nil
}
