package promotion_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/src/modules/promotion_module/dtos"
)

// GetPromotions lists promotions, optionally filtered to one store
//
//	@Summary		List promotions
//	@Description	Get promotions, optionally filtered by storeId
//	@Tags			Promotion Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			storeId	query		int	false	"store id"
//	@Success		200		{array}		responses.PromotionListItem
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/promotions [get]
func (c Controller) GetPromotions(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetPromotionsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPromotions")))

	storeID := f.QueryInt("storeId", 0)

	responseData, err := c.promotionService().GetPromotions(ctx, storeID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreatePromotion creates a store-wide promotion with per-product prices
//
//	@Summary		Create promotion
//	@Description	Create a promotion for one store with special prices for one or
//	@Description	more products. Rejected if its time range overlaps an existing
//	@Description	promotion for the same store.
//	@Tags			Promotion Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.CreatePromotion	true	"promotion"
//	@Success		200		{object}	responses.PromotionDetailResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/promotions [post]
func (c Controller) CreatePromotion(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "CreatePromotionController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreatePromotion")))

	dto := new(dtos.CreatePromotion)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	responseData, err := c.promotionService().CreatePromotion(ctx, dto)
	if err != nil {
		if err == exception.ErrPromotionOverlap {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.PromotionOverlapResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
