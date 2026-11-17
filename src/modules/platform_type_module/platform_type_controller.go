package platform_type_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

// GetPlatformTypes lists the active sales platforms
//
//	@Summary		List platform types
//	@Description	Get every active sales platform a receipt bill can be tagged with (id, code, name, logoUrl), in display order.
//	@Tags			Platform Type Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		responses.PlatformTypeResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Security		BearerAuth
//	@Router			/api/v1/platform-types [get]
func (c Controller) GetPlatformTypes(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetPlatformTypesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetPlatformTypes")))

	responseData, err := c.platformTypeService().GetPlatformTypes(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
