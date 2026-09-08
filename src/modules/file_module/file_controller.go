package file_module

import (
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

// GetFile streams a previously uploaded file (e.g. a bill slip) by its
// storage key.
//
//	@Summary		Get an uploaded file
//	@Description	Stream a file previously uploaded via POST/PUT /bills (e.g. a
//	@Description	payment slip) by its storage key, such as bills/slips/<uuid>.jpeg
//	@Tags			File Module (Version 1)
//	@Produce		application/octet-stream
//	@Param			path	path	string	true	"storage key, e.g. bills/slips/<uuid>.jpeg"
//	@Success		200
//	@Failure		400	{object}	exception.ErrorResponse
//	@Failure		404	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/files/{path} [get]
func (c Controller) GetFile(f *fiber.Ctx) error {
	var (
		key       = f.Params("*")
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetFileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetFile"), attribute.String("key", key)))
	)

	// Reject path traversal outright — this key ends up in a filesystem/S3
	// object read, never validated against a known set of files.
	if key == "" || strings.Contains(key, "..") {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter)
	}

	file, err := c.fileService().GetFile(ctx, key)
	if err != nil {
		if os.IsNotExist(err) {
			return exception.HttpErrorResponseMapping(f, fiber.StatusNotFound, exception.RecordNotFoundResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}
	defer file.Close()

	if contentType := mime.TypeByExtension(filepath.Ext(key)); contentType != "" {
		f.Set(fiber.HeaderContentType, contentType)
	}

	c.m.tracer.TraceEnd(span)
	return f.SendStream(file)
}
