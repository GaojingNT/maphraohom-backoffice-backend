package store_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
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

// GetStoreProducts lists every product's currently effective price at a store
//
//	@Summary		List a store's product prices
//	@Description	Get every product's currently effective price at this store
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"store id"
//	@Success		200	{array}		responses.StoreProductPriceItem
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id}/products [get]
func (c Controller) GetStoreProducts(f *fiber.Ctx) error {
	storeID, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetStoreProductsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetStoreProducts"), attribute.Int("storeId", storeID)))

	responseData, err := c.storeService().GetStoreProducts(ctx, storeID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetStoreBasePrices lists a store's editable base prices, never resolved
// against an active promotion
//
//	@Summary		List a store's editable base prices
//	@Description	Get every product's raw base price at this store (ignores active promotions) — for the price-management admin screen
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"store id"
//	@Success		200	{array}		responses.StoreProductBasePriceItem
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id}/base-prices [get]
func (c Controller) GetStoreBasePrices(f *fiber.Ctx) error {
	storeID, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetStoreBasePricesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetStoreBasePrices"), attribute.Int("storeId", storeID)))

	responseData, err := c.storeService().GetStoreBasePrices(ctx, storeID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// UpdateStoreProductPrice sets a store's base price for one product
//
//	@Summary		Update a store's base price for one product
//	@Description	Set the store's base price for one product (upserts the (store, product) row — no history is kept)
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id			path		int								true	"store id"
//	@Param			productId	path		int								true	"product id"
//	@Param			body		body		dtos.UpdateStoreProductPrice	true	"price"
//	@Success		200			{object}	responses.StoreProductBasePriceItem
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id}/products/{productId} [put]
func (c Controller) UpdateStoreProductPrice(f *fiber.Ctx) error {
	storeID, _ := f.ParamsInt("id")
	productID, _ := f.ParamsInt("productId")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "UpdateStoreProductPriceController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateStoreProductPrice"), attribute.Int("storeId", storeID), attribute.Int("productId", productID)))

	dto := new(dtos.UpdateStoreProductPrice)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	responseData, err := c.storeService().UpdateStoreProductPrice(ctx, storeID, productID, dto.Price)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetStoreProduct gets one product's currently effective price at a store
//
//	@Summary		Get a store's price for one product
//	@Description	Get a single product's currently effective price at this store
//	@Tags			Store Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id			path		int	true	"store id"
//	@Param			productId	path		int	true	"product id"
//	@Success		200			{object}	responses.StoreProductPriceItem
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/stores/{id}/products/{productId} [get]
func (c Controller) GetStoreProduct(f *fiber.Ctx) error {
	storeID, _ := f.ParamsInt("id")
	productID, _ := f.ParamsInt("productId")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetStoreProductController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetStoreProduct"), attribute.Int("storeId", storeID), attribute.Int("productId", productID)))

	responseData, err := c.storeService().GetStoreProduct(ctx, storeID, productID)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
