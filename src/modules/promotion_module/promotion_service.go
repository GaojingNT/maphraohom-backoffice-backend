package promotion_module

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/src/modules/promotion_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/promotion_module/responses"
)

func (s Service) GetPromotions(ctx context.Context, storeID int) ([]responses.PromotionListItem, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "GetPromotionsService", trace.WithAttributes(attribute.String("service", "GetPromotions")))

	promotions, err := s.promotionRepository().GetPromotions(ctx, storeID)

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return responses.PromotionListItem{}.Collection(promotions), nil
}

func (s Service) CreatePromotion(ctx context.Context, dto *dtos.CreatePromotion) (*responses.PromotionDetailResponse, error) {
	ctx, childSpan := s.tracer.TraceStart(ctx, "CreatePromotionService", trace.WithAttributes(attribute.String("service", "CreatePromotion")))

	items := make([]CreatePromotionItemInput, 0, len(dto.Items))
	for _, item := range dto.Items {
		items = append(items, CreatePromotionItemInput{
			ProductID: item.ProductID,
			Price:     item.Price,
		})
	}

	promotion, err := s.promotionRepository().CreatePromotion(ctx, CreatePromotionInput{
		Name:     dto.Name,
		StoreID:  dto.StoreID,
		StartsAt: dto.StartsAt,
		EndsAt:   dto.EndsAt,
		Items:    items,
	})

	s.tracer.TraceEnd(childSpan)

	if err != nil {
		return nil, err
	}

	return new(responses.PromotionDetailResponse).Make(promotion), nil
}
