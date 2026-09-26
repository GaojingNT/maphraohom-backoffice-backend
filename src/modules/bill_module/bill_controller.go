package bill_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/http_response"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/bill_module/responses"
)

// GetBills lists all existing bills
//
//	@Summary		List bills
//	@Description	Get bills (paginated) — id, type, storeId, storeName, bookNo,
//	@Description	receiptNo, customerName, customerAddress, total, itemCount, hasSlip, createdAt
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Param			search	query		string	false	"search keyword"
//	@Param			type	query		string	false	"filter by bill type: receipt or payment"
//	@Param			storeId	query		int		false	"filter by store id"
//	@Param			from	query		string	false	"filter created_at >= this date, YYYY-MM-DD"
//	@Param			to		query		string	false	"filter created_at <= this date, YYYY-MM-DD"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		400		{object}	exception.ErrorResponse
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
	queryType := f.Query("type")
	queryStoreID := f.QueryInt("storeId", 0)
	queryFrom := f.Query("from")
	queryTo := f.Query("to")

	if queryType != "" && !models.IsValidBillType(queryType) {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, exception.ParameterError{
			FailedField: "type",
			Tag:         "oneof=receipt payment",
			Value:       queryType,
		})
	}

	// Get paginate values
	// Newest first by the bill's own date (created_at can be back-dated, so
	// id order is not date order); id breaks ties. Paginate renders this as
	// "created_at desc, id desc".
	paginate := paginator.NewPagination(
		paginator.WithPage(queryPage),
		paginator.WithLimit(queryLimit),
		paginator.WithOrderBy("created_at desc, id"),
		paginator.WithSort("desc"),
		paginator.WithAttributes("search", querySearch),
		paginator.WithAttributes("search_by", querySearchBy),
		paginator.WithAttributes("searchable", models.BillSearchable()),
		paginator.WithAttributes("type", queryType),
		paginator.WithAttributes("store_id", queryStoreID),
		paginator.WithAttributes("from", queryFrom),
		paginator.WithAttributes("to", queryTo),
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
//	@Description	Get bill by id — every bills column except deleted_at, plus a ready-to-use slipUrl
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
//	@Description	Create a bill (JSON) with one or more line items, priced by the
//	@Description	caller (no price list). Book/receipt numbers are computed
//	@Description	server-side, scoped per (store, type); unit/subtotal/total and
//	@Description	anything else derivable are ignored if the client sends them.
//	@Description	createdAt is optional (RFC3339) to back-/post-date the bill —
//	@Description	defaults to now when omitted; never affects book/receipt numbering.
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.CreateBill	true	"bill"
//	@Success		201		{object}	responses.BillDetailResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/bills [post]
func (c Controller) CreateBill(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateBillController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateBill")))
		err       error
	)

	dto := new(dtos.CreateBill)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// JwtAuthProtected guards every bill route, so authUser is always set.
	authUser := f.Locals("authUser").(*models.User)

	responseData, fieldErrors, err := c.billService().CreateBill(ctx, dto, authUser.ID)
	if len(fieldErrors) > 0 {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, fieldErrors...)
	}
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusCreated).JSON(responseData)
}

// UpdateBill replaces an existing bill's editable fields and line items
//
//	@Summary		Update bill
//	@Description	Replace a bill's customer, items, discount, and shipping fee
//	@Description	(JSON, same shape as create). storeId and type must match the
//	@Description	existing bill — sending different values is rejected with 400.
//	@Description	Book/receipt numbers and the slip are never touched here.
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int				true	"bill id"
//	@Param			body	body		dtos.UpdateBill	true	"bill"
//	@Success		200		{object}	responses.BillDetailResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/bills/{id} [put]
func (c Controller) UpdateBill(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UpdateBillController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateBill"), attribute.Int("id", id)))
		err       error
	)

	dto := new(dtos.UpdateBill)
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	responseData, fieldErrors, err := c.billService().UpdateBill(ctx, id, dto)
	if len(fieldErrors) > 0 {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, fieldErrors...)
	}
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		if err == exception.ErrBillFieldImmutable {
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.BillFieldImmutableResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// DeleteBill removes an existing bill by ID (soft delete)
//
//	@Summary		Delete bill
//	@Description	Delete bill by id — soft delete only, slip and book/receipt number are kept as-is (never reused)
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

// UploadSlip attaches (or replaces) a bill's payment slip
//
//	@Summary		Upload a bill's slip
//	@Description	Upload a slip image (jpeg/png/webp, checked by magic bytes, ≤ 10MB) for an existing, non-deleted bill. Replaces any existing slip.
//	@Tags			Bill Module (Version 1)
//	@Accept			mpfd
//	@Produce		json
//	@Param			id		path		int		true	"bill id"
//	@Param			slip	formData	file	true	"slip image"
//	@Success		200		{object}	object{slipUrl=string}
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/bills/{id}/slip [put]
func (c Controller) UploadSlip(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UploadSlipController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UploadSlip"), attribute.Int("id", id)))
	)

	file, err := f.FormFile("slip")
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, exception.ParameterError{
			FailedField: "slip",
			Tag:         "required",
			Value:       "",
		})
	}

	slipURL, err := c.billService().UploadSlip(ctx, id, file)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		case exception.ErrUnsupportedSlipType:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.UnsupportedSlipTypeResponseError, err)
		case exception.ErrSlipFileTooLarge:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.SlipFileTooLargeResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{"slipUrl": slipURL})
}

// DeleteSlip removes a bill's slip
//
//	@Summary		Delete a bill's slip
//	@Description	Clear a bill's slip and remove the underlying file (best-effort)
//	@Tags			Bill Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"bill id"
//	@Success		200	{object}	http_response.OkResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/bills/{id}/slip [delete]
func (c Controller) DeleteSlip(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeleteSlipController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteSlip"), attribute.Int("id", id)))
	)

	if err := c.billService().DeleteSlip(ctx, id); err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Slip deleted successfully")
}
