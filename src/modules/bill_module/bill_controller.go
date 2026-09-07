package bill_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/responses"
)

// GetBills lists all existing bills
//
//	@Summary		List bills
//	@Description	Get bills (paginated) — id, customerName, customerAddress, total
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Param			search	query		string	false	"search keyword"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/bills [get]
func (c Controller) GetBills(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetBillsController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetBills")))
		responseData *paginator.Pagination
		err          error
	)

	// Get query strings
	queryPage := f.QueryInt("page", 1)
	queryLimit := f.QueryInt("limit", 20)
	querySearch := f.Query("search")
	querySearchBy := f.Query("searchBy")

	// Get paginate values
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithAttributes("search", querySearch),
		paginator.WithAttributes("search_by", querySearchBy),
		paginator.WithAttributes("searchable", models.BillSearchable()),
	)

	responseData, err = c.billService().GetBills(ctx, paginate)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetBill get existing bill by ID
//
//	@Summary		Get bill
//	@Description	Get bill by id — every bills column except deleted_at
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"bill id"
//	@Success		200	{object}	responses.BillDetailResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/bills/{id} [get]
func (c Controller) GetBill(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetBillController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetBill"), attribute.Int("id", id)))
		responseData *responses.BillDetailResponse
		err          error
	)

	responseData, err = c.billService().GetBill(ctx, id)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateBill creates a new bill
//
//	@Summary		Create bill
//	@Description	Create a bill. Book/receipt numbers and price are computed
//	@Description	server-side; the price is looked up from the store's current
//	@Description	product price and the receipt number resets every 50 receipts
//	@Description	(new book) and every calendar year (back to book 1 / receipt 1).
//	@Tags			Bill Module (Version 1)
//	@Accept			mpfd
//	@Produce		json
//	@Param			productId		formData	int		true	"product id"
//	@Param			storeId			formData	int		true	"store id"
//	@Param			customerName	formData	string	true	"customer name"
//	@Param			customerAddress	formData	string	true	"customer address"
//	@Param			kilogram		formData	number	true	"kilogram"
//	@Param			discount		formData	number	false	"discount"
//	@Param			shippingFee		formData	number	false	"shipping fee"
//	@Param			slip			formData	file	false	"slip image"
//	@Success		200				{object}	responses.BillDetailResponse
//	@Failure		400				{object}	exception.ErrorResponse
//	@Failure		500				{object}	exception.ErrorResponse
//	@Router			/api/v1/bills [post]
func (c Controller) CreateBill(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateBillController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateBill")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.CreateBill)

	// Parse HTTP request body to struct variable
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Slip is optional — ignore the error when the field is simply absent
	slip, _ := f.FormFile("slip")

	responseData, err := c.billService().CreateBill(ctx, dto, slip)
	if err != nil {
		if err == exception.ErrPriceNotConfigured {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.PriceNotConfiguredResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
