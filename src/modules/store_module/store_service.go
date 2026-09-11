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

// GetStoreBasePrices lists a store's editable base prices (never resolved
// against an active promotion) — used by the price-management admin screen.
func (s Service) GetStoreBasePrices(ctx context.Context, storeID int) ([]responses.StoreProductBasePriceItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetStoreBasePricesService", trace.WithAttributes(attribute.String("service", "GetStoreBasePrices")))

	prices, err := s.storeRepository().GetStoreProducts(ctx, storeID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return responses.StoreProductBasePriceItem{}.Collection(prices), nil
}

// UpdateStoreProductPrice sets a store's base price for one product.
func (s Service) UpdateStoreProductPrice(ctx context.Context, storeID int, productID int, price float64) (*responses.StoreProductBasePriceItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "UpdateStoreProductPriceService", trace.WithAttributes(attribute.String("service", "UpdateStoreProductPrice")))

	record, err := s.storeRepository().UpdateStoreProductPrice(ctx, storeID, productID, price)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	item := responses.StoreProductBasePriceItem{}.Make(record)
	return &item, nil
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
