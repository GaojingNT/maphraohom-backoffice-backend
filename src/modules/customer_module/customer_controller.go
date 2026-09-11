package customer_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/customer_module/dtos"
)

// GetCustomers lists all existing customers
//
//	@Summary		List customers
//	@Description	Get customers (paginated)
//	@Tags			Customer Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Param			search	query		string	false	"search keyword"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/customers [get]
func (c Controller) GetCustomers(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetCustomersController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetCustomers")))
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
		paginator.WithAttributes("searchable", models.CustomerSearchable()),
	)

	responseData, err = c.customerService().GetCustomers(ctx, paginate)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetCustomer get existing customer by ID
//
//	@Summary		Get customer
//	@Description	Get customer by id
//	@Tags			Customer Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"customer id"
//	@Success		200	{object}	responses.CustomerDetailResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/customers/{id} [get]
func (c Controller) GetCustomer(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetCustomerController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetCustomer"), attribute.Int("id", id)))

	responseData, err := c.customerService().GetCustomer(ctx, id)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetCustomerAddresses lists a customer's addresses
//
//	@Summary		List a customer's addresses
//	@Description	Get every address on file for this customer
//	@Tags			Customer Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"customer id"
//	@Success		200	{array}		responses.CustomerAddressItem
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/customers/{id}/addresses [get]
func (c Controller) GetCustomerAddresses(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetCustomerAddressesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetCustomerAddresses"), attribute.Int("id", id)))

	responseData, err := c.customerService().GetCustomerAddresses(ctx, id)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetCustomerPhones lists a customer's phone numbers
//
//	@Summary		List a customer's phone numbers
//	@Description	Get every phone number on file for this customer (default-first)
//	@Tags			Customer Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"customer id"
//	@Success		200	{array}		responses.CustomerPhoneItem
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/customers/{id}/phones [get]
func (c Controller) GetCustomerPhones(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "GetCustomerPhonesController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetCustomerPhones"), attribute.Int("id", id)))

	responseData, err := c.customerService().GetCustomerPhones(ctx, id)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateCustomer creates a new customer
//
//	@Summary		Create customer
//	@Description	Create a new customer by name
//	@Tags			Customer Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.CreateCustomer	true	"customer"
//	@Success		200		{object}	responses.CustomerDetailResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/customers [post]
func (c Controller) CreateCustomer(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "CreateCustomerController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateCustomer")))

	dto := new(dtos.CreateCustomer)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	responseData, err := c.customerService().CreateCustomer(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateCustomerAddress adds a new address for a customer
//
//	@Summary		Add a customer address
//	@Description	Add a new address for a customer, optionally marking it default
//	@Tags			Customer Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"customer id"
//	@Param			body	body		dtos.CreateCustomerAddress	true	"address"
//	@Success		200		{object}	responses.CustomerAddressItem
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/customers/{id}/addresses [post]
func (c Controller) CreateCustomerAddress(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "CreateCustomerAddressController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateCustomerAddress"), attribute.Int("id", id)))

	dto := new(dtos.CreateCustomerAddress)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	responseData, err := c.customerService().CreateCustomerAddress(ctx, id, dto)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateCustomerPhone adds a new phone number for a customer
//
//	@Summary		Add a customer phone number
//	@Description	Add a new phone number for a customer, optionally marking it default
//	@Tags			Customer Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"customer id"
//	@Param			body	body		dtos.CreateCustomerPhone	true	"phone"
//	@Success		200		{object}	responses.CustomerPhoneItem
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/customers/{id}/phones [post]
func (c Controller) CreateCustomerPhone(f *fiber.Ctx) error {
	id, _ := f.ParamsInt("id")
	ctx, span := c.m.tracer.TraceStart(f.Context(), "CreateCustomerPhoneController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateCustomerPhone"), attribute.Int("id", id)))

	dto := new(dtos.CreateCustomerPhone)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	responseData, err := c.customerService().CreateCustomerPhone(ctx, id, dto)
	if err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}
