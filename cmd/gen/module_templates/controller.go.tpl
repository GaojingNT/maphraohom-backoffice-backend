package {{.ModuleName}}_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/pkg/database/paginator"
	"maphraohom.app/maphraohom-backoffice/internal/utils"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/{{.ModuleName}}_module/dtos"
)

type IController interface {
	Get{{.PascalModulePluralName}}(f *fiber.Ctx) error
	Get{{.PascalModuleName}}(f *fiber.Ctx) error
	Create{{.PascalModuleName}}(f *fiber.Ctx) error
	Update{{.PascalModuleName}}(f *fiber.Ctx) error
	Delete{{.PascalModuleName}}(f *fiber.Ctx) error
}

// Get{{.PascalModulePluralName}} lists all existing {{.ModulePluralName}}
//
//	@Summary		List {{.ModulePluralName}}
//	@Description	get {{.ModulePluralName}}
//	@Tags			{{.PascalModuleName}} Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			page	query		string	false	"page number"
//	@Param			limit	query		string	false	"page size"
//	@Param			search	query		string	false	"page size"
//	@Success		200		{object}	paginator.Pagination
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/{{.ModulePluralName}} [get]
func (c Controller) Get{{.PascalModulePluralName}}(f *fiber.Ctx) error {
	var (
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "Get{{.PascalModulePluralName}}Controller", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "Get{{.PascalModulePluralName}}")))
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
		paginator.WithAttributes("searchable", models.{{.PascalModuleName}}Searchable()),
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"{{.ModulePluralName}}"}, "Get{{.PascalModulePluralName}}", paginate.Page, paginate.Limit, paginate.Attributes)

	// Get data from cache
	responseData, err = c.m.cacher.PaginationCache(ctx, cacheKey, cacheTags, paginate, c.{{.CamelModuleName}}Service().Get{{.PascalModulePluralName}})
	if err != nil {
		return exception.HttpErrorResponseMapping(f, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// Get{{.PascalModuleName}} get existing {{.ModuleName}} by ID
//
//	@Summary		Get {{.ModuleName}}
//	@Description	Get {{.ModuleName}}
//	@Tags			{{.PascalModuleName}} Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			{{.ModuleName}}_id	path		int	true	"{{.ModuleName}} id"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		404		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/{{.ModulePluralName}}/{{.BracketModuleNameID}} [get]
func (c Controller) Get{{.PascalModuleName}}(f *fiber.Ctx) error {
	var (
		id, _        = f.ParamsInt("id")
		ctx, span    = c.m.tracer.TraceStart(f.Context(), "Get{{.PascalModuleName}}Controller", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "Get{{.PascalModuleName}}"), attribute.Int("id", id)))
		responseData map[string]interface{}
		err          error
	)

	// Make cache indexing
	cacheTags, cacheKey := c.m.cacher.Indexing([]string{"{{.ModulePluralName}}"}, "Get{{.PascalModuleName}}", id)

	// Get data from cache
	responseData, err = c.m.cacher.QueryByIntParamCache(ctx, cacheKey, cacheTags, id, c.{{.CamelModuleName}}Service().Get{{.PascalModuleName}})
	if err != nil {
		return exception.HttpErrorResponseMapping(f, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(responseData)
}

// Create{{.PascalModuleName}} create new {{.ModuleName}}
//
//	@Summary		Create {{.ModuleName}}
//	@Description	Create an new {{.ModuleName}}
//	@Tags			{{.PascalModuleName}} Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			firstName	body		string	true	"{{.ModuleName}} first name"
//	@Param			lastName	body		string	true	"{{.ModuleName}} last name"
//	@Param			email		body		string	true	"{{.ModuleName}} email"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		404			{object}	http_response.OkResponse
//	@Failure		500			{object}	http_response.OkResponse
//	@Router			/api/v1/{{.ModulePluralName}} [post]
func (c Controller) Create{{.PascalModuleName}}(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "Create{{.PascalModuleName}}Controller", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "Create{{.PascalModuleName}}")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.Create{{.PascalModuleName}})

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.{{.CamelModuleName}}Service().Create{{.PascalModuleName}}(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, err)
	}

	// Clear {{.ModuleName}} cache
	c.m.cacher.Tag("{{.ModulePluralName}}").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// Update{{.PascalModuleName}} update existing {{.ModuleName}}
//
//	@Summary		Update {{.ModuleName}}
//	@Description	Update an existing {{.ModuleName}}
//	@Tags			{{.PascalModuleName}} Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			{{.ModuleName}}_id		path		int		true	"{{.ModuleName}} id"
//	@Param			firstName	body		string	true	"{{.ModuleName}} first name"
//	@Param			lastName	body		string	true	"{{.ModuleName}} last name"
//	@Param			email		body		string	true	"{{.ModuleName}} email"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		404			{object}	http_response.OkResponse
//	@Failure		500			{object}	http_response.OkResponse
//	@Router			/api/v1/{{.ModulePluralName}}/{{.BracketModuleNameID}} [put]
func (c Controller) Update{{.PascalModuleName}}(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "Update{{.PascalModuleName}}Controller", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "Update{{.PascalModuleName}}"), attribute.Int("id", id)))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.Update{{.PascalModuleName}})

	// Parse HTTP request body to struct variable
	if err = f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.{{.CamelModuleName}}Service().Update{{.PascalModuleName}}(ctx, id, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, err)
	}

	// Clear {{.ModuleName}} cache
	c.m.cacher.Tag("{{.ModulePluralName}}").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// Delete{{.PascalModuleName}} remove existing {{.ModuleName}} by ID
//
//	@Summary		Delete {{.ModuleName}}
//	@Description	Delete {{.ModuleName}}
//	@Tags			{{.PascalModuleName}} Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			{{.ModuleName}}_id	path		int	true	"{{.ModuleName}} id"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		404		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/{{.ModulePluralName}}/{{.BracketModuleNameID}} [delete]
func (c Controller) Delete{{.PascalModuleName}}(f *fiber.Ctx) error {
	var (
		id, _     = f.ParamsInt("id")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "Delete{{.PascalModuleName}}Controller", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "Delete{{.PascalModuleName}}"), attribute.Int("id", id)))
		err       error
	)

	// Call service function
	err = c.{{.CamelModuleName}}Service().Delete{{.PascalModuleName}}(ctx, id)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, err)
	}

	// Clear {{.ModuleName}} cache
	c.m.cacher.Tag("{{.ModulePluralName}}").Flush(ctx)

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}
