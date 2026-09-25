package store_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/http_response"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/store_module/dtos"
)

// GetStores lists all existing stores
//
//	@Summary		List stores
//	@Description	Get stores (paginated)
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Param			search	query		string	false	"search keyword"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/stores [get]
func (c Controller) GetStores(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetStoresController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetStores")))
		responseData *paginator.Pagination
		err          error
	)

	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)
	querySearch := f.Query("search")
	querySearchBy := f.Query("searchBy")

	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithAttributes("search", querySearch),
		paginator.WithAttributes("search_by", querySearchBy),
		paginator.WithAttributes("searchable", models.StoreSearchable()),
	)

	responseData, err = c.storeService().GetStores(ctx, paginate)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetStore get existing store by ID
//
//	@Summary		Get store
//	@Description	Get store by id
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"store id"
//	@Success		200	{object}	responses.StoreDetailResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id} [get]
func (c Controller) GetStore(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetStoreController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetStore"), attribute.Int("id", id)))

	responseData, err := c.storeService().GetStore(ctx, id)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// UpdateStore replaces a store's name, address, and phone
//
//	@Summary		Update store
//	@Description	Replace a store's name, address, and phone. The logo is uploaded through its own endpoint.
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"store id"
//	@Param			body	body		dtos.UpdateStore	true	"store"
//	@Success		200		{object}	responses.StoreDetailResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id} [put]
func (c Controller) UpdateStore(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "UpdateStoreController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateStore"), attribute.Int("id", id)))

	dto := new(dtos.UpdateStore)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	responseData, err := c.storeService().UpdateStore(ctx, id, dto)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// UploadLogo attaches (or replaces) a store's logo
//
//	@Summary		Upload a store's logo
//	@Description	Upload a logo image (jpeg/png/webp, checked by magic bytes, ≤ 10MB) for an existing store. Replaces any existing logo.
//	@Tags			Store Module (Version 1)
//	@Accept			mpfd
//	@Produce		json
//	@Param			id		path		int		true	"store id"
//	@Param			logo	formData	file	true	"logo image"
//	@Success		200		{object}	object{logo=string}
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id}/logo [put]
func (c Controller) UploadLogo(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "UploadLogoController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UploadLogo"), attribute.Int("id", id)))

	file, err := f.FormFile("logo")
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, exception.ParameterError{
			FailedField: "logo",
			Tag:         "required",
			Value:       "",
		})
	}

	logoURL, err := c.storeService().UploadLogo(ctx, id, file)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrUnsupportedImageType:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.UnsupportedImageTypeResponseError, err)
		case exception.ErrImageFileTooLarge:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.ImageFileTooLargeResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{"logo": logoURL})
}

// DeleteLogo removes a store's logo
//
//	@Summary		Delete a store's logo
//	@Description	Clear a store's logo and remove the underlying file (best-effort)
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"store id"
//	@Success		200	{object}	http_response.OkResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id}/logo [delete]
func (c Controller) DeleteLogo(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "DeleteLogoController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteLogo"), attribute.Int("id", id)))

	if err := c.storeService().DeleteLogo(ctx, id); err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Logo deleted successfully")
}

// GetLastPrices lists, per product, the price last used at this store for a
// given bill type
//
//	@Summary		List a store's last-used prices
//	@Description	Get, for every product, the price used in this store's most recent bill of the given type — used to prefill the create-bill form (optional).
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			storeId	path		int		true	"store id"
//	@Param			type	query		string	true	"bill type: receipt or payment"
//	@Success		200		{array}		responses.LastPriceItem
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{storeId}/last-prices [get]
func (c Controller) GetLastPrices(f *fiber.Ctx) error {
	storeID, _ := f.ParamsInt("storeId")
	billType := f.Query("type")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetLastPricesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetLastPrices"), attribute.Int("storeId", storeID), attribute.String("type", billType)))

	if !models.IsValidBillType(billType) {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, exception.ParameterError{
			FailedField: "type",
			Tag:         "oneof=receipt payment",
			Value:       billType,
		})
	}

	responseData, err := c.storeService().GetLastPrices(ctx, storeID, billType)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
