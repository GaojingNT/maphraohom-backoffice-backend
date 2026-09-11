package product_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

// GetProducts lists all existing products
//
//	@Summary		List products
//	@Description	Get every product (id, name, unit)
//	@Tags			Product Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		responses.ProductListItem
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/products [get]
func (c Controller) GetProducts(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetProductsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetProducts")))

	responseData, err := c.productService().GetProducts(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
