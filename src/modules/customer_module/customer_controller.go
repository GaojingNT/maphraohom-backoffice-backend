package customer_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
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
