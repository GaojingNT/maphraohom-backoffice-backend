package bill_module

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/http_response"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/responses"
)

// GetBills lists all existing bills
//
//	@Summary		List bills
//	@Description	Get bills (paginated) — id, storeId, storeName, receiptNo,
//	@Description	customerName, customerAddress, total, totalKilogram, itemCount, createdAt
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Param			search	query		string	false	"search keyword"
//	@Param			period	query		string	false	"filter by created_at: day, week, month, or year"
//	@Param			date	query		string	false	"reference date for period, YYYY-MM-DD (default: today)"
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
	queryPeriod := f.Query("period")
	queryDate := f.Query("date")

	// Get paginate values
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithAttributes("search", querySearch),
		paginator.WithAttributes("search_by", querySearchBy),
		paginator.WithAttributes("searchable", models.BillSearchable()),
		paginator.WithAttributes("period", queryPeriod),
		paginator.WithAttributes("date", queryDate),
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
//	@Description	Create a bill with one or more line items. Book/receipt
//	@Description	numbers and each item's price are computed server-side (price
//	@Description	looked up from the store's current product price); the receipt
//	@Description	number resets every 50 receipts (new book) and every calendar
//	@Description	year (back to book 1 / receipt 1).
//	@Tags			Bill Module (Version 1)
//	@Accept			mpfd
//	@Produce		json
//	@Param			storeId			formData	int		true	"store id"
//	@Param			customerName	formData	string	true	"customer name"
//	@Param			customerAddress	formData	string	true	"customer address"
//	@Param			items			formData	string	true	"JSON array, e.g. [{\"productId\":1,\"kilogram\":2.5}]"
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

	// Items is a JSON-encoded array within the multipart form (see dtos.CreateBill).
	var items []dtos.CreateBillItem
	if err = json.Unmarshal([]byte(dto.Items), &items); err != nil || len(items) == 0 {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}
	for _, item := range items {
		if itemErrors := validator.Validate(item); itemErrors != nil {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, itemErrors...)
		}
	}

	// Slip is optional — ignore the error when the field is simply absent
	slip, _ := f.FormFile("slip")

	responseData, err := c.billService().CreateBill(ctx, dto, items, slip)
	if err != nil {
		if err == exception.ErrPriceNotConfigured {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.PriceNotConfiguredResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// DeleteBill removes an existing bill by ID (soft delete)
//
//	@Summary		Delete bill
//	@Description	Delete bill by id
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"bill id"
//	@Success		200	{object}	http_response.OkResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/bills/{id} [delete]
func (c Controller) DeleteBill(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeleteBillController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteBill"), attribute.Int("id", id)))
		err       error
	)

	err = c.billService().DeleteBill(ctx, id)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Bill deleted successfully")
}
