package healthcheck_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

// CheckDatabaseConnection check the database connection status
//
//	@Summary		Healthcheck
//	@Description	check the API status
//	@Tags			Healthcheck
//	@Produce		html
//	@Success		200	{object}	string
//	@Failure		500	{object}	string
//	@Router			/health [get]
func (c Controller) CheckDatabaseConnection(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "CheckDatabaseConnectionController", trace.WithAttributes(attribute.String("controller", "CheckDatabaseConnection")))
		err       error
	)

	// Call service function
	err = c.dbRepository().CheckDatabaseConnection(ctx)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.SendStatus(fiber.StatusOK)
}

// // Note: This function is not used in the project, for example purposes only
// func (c Controller) TestUploadFile(f *fiber.Ctx) error {
// 	var (
// 		ctx, span = c.tracer.TraceStart(f.Context(), "TestUploadFileController", trace.WithAttributes(attribute.String("controller", "TestUploadFile")))
// 		err       error
// 	)

// 	// Parse file from request
// 	fh, err := f.FormFile("file")
// 	if err != nil {
// 		return exception.HttpErrorResponseMapping(f, err)
// 	}

// 	// Call service function
// 	err = c.dbRepository().TestUploadFile(ctx, fh)
// 	if err != nil {
// 		return exception.HttpErrorResponseMapping(f, err)
// 	}

// 	c.tracer.TraceEnd(span)
// 	return f.SendStatus(fiber.StatusOK)
// }
