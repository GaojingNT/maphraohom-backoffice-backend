package user_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/http_response"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/pkg/cache"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/user_module/dtos"
	"maphraohom.app/maphraohom-backoffice/src/modules/user_module/responses"
)

// GetUsers lists all existing users
//
//	@Summary		List users
//	@Description	get users
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Param			search	query		string	false	"page size"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/users [get]
func (c Controller) GetUsers(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetUsersController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetUsers")))
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
		paginator.WithAttributes("searchable", models.UserSearchable()),
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"users"}, "GetUsers", paginate.Page, paginate.Limit, paginate.Attributes)

	// Get data from cache
	responseData, err = c.m.cacher.PaginationCache(ctx, cacheKey, cacheTags, paginate, c.userService().GetUsers)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// GetUser get existing user by ID
//
//	@Summary		Get user
//	@Description	Get user
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		int	true	"user id"
//	@Success		200		{object}	responses.GetUserByIDResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id} [get]
func (c Controller) GetUser(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "GetUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetUser"), attribute.Int("id", id)))
		responseData *responses.GetUserByIDResponse
		err          error
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"users"}, "GetUser", id)

	// Get data from cache
	responseData, err = cache.QueryByParamCache(c.m.cacher, ctx, cacheKey, cacheTags, id, c.userService().GetUser)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// CreateUser create new user
//
//	@Summary		Create user
//	@Description	Create an new user
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			firstName	body		string	true	"user first name"
//	@Param			lastName	body		string	true	"user last name"
//	@Param			email		body		string	true	"user email"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/users [post]
func (c Controller) CreateUser(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CreateUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "CreateUser")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.CreateUser)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.userService().CreateUser(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear user cache
	c.m.cacher.Tag("users").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "User created successfully")
}

// UpdateUser update existing user
//
//	@Summary		Update user
//	@Description	Update an existing user
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			user_id		path		int		true	"user id"
//	@Param			firstName	body		string	true	"user first name"
//	@Param			lastName	body		string	true	"user last name"
//	@Param			email		body		string	true	"user email"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		404			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id} [put]
func (c Controller) UpdateUser(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "UpdateUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateUser"), attribute.Int("id", id)))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.UpdateUser)

	// Parse HTTP request body to struct variable
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.userService().UpdateUser(ctx, id, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear user cache
	c.m.cacher.Tag("users").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "User updated successfully")
}

// DeleteUser remove existing user by ID
//
//	@Summary		Delete user
//	@Description	Delete user
//	@Tags			User Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		int	true	"user id"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		404		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/users/{user_id} [delete]
func (c Controller) DeleteUser(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "DeleteUserController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteUser"), attribute.Int("id", id)))
		err       error
	)

	// Call service function
	err = c.userService().DeleteUser(ctx, id)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	// Clear user cache
	c.m.cacher.Tag("users").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "User deleted successfully")
}
