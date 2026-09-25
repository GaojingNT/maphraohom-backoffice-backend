package auth_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/http_response"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/auth_module/dtos"
)

// UpdateProfile updates the signed-in user's own profile
//
//	@Summary		Update profile
//	@Description	Replace the signed-in user's first name, last name and email. The signature is uploaded through its own endpoint.
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dtos.UpdateProfileDto	true	"profile"
//	@Success		200		{object}	responses.GetProfileResponse
//	@Failure		400		{object}	exception.ErrorResponse
//	@Failure		401		{object}	exception.ErrorResponse
//	@Failure		409		{object}	exception.ErrorResponse
//	@Failure		500		{object}	exception.ErrorResponse
//	@Router			/api/v1/auth/profile [put]
func (c Controller) UpdateProfile(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "UpdateProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UpdateProfile")))

	user := f.Locals("authUser").(*models.User)

	dto := new(dtos.UpdateProfileDto)
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	if errors := validator.Validate(*dto); errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	response, err := c.authService().UpdateProfile(ctx, user.ID, dto)
	if err != nil {
		switch err {
		case exception.ErrEmailAlreadyTaken:
			return exception.HttpErrorResponseMapping(f, fiber.StatusConflict, exception.EmailAlreadyTakenResponseError, err)
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(response)
}

// UploadSignature attaches (or replaces) the signed-in user's signature
//
//	@Summary		Upload my signature
//	@Description	Upload a signature image (jpeg/png/webp, checked by magic bytes, ≤ 10MB) for the signed-in user. Replaces any existing signature. Printed on every receipt this user exports.
//	@Tags			Auth Module (Version 1)
//	@Accept			mpfd
//	@Produce		json
//	@Param			signature	formData	file	true	"signature image"
//	@Success		200			{object}	object{signature=string}
//	@Failure		400			{object}	exception.ErrorResponse
//	@Failure		401			{object}	exception.ErrorResponse
//	@Failure		500			{object}	exception.ErrorResponse
//	@Router			/api/v1/auth/profile/signature [put]
func (c Controller) UploadSignature(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "UploadSignatureController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "UploadSignature")))

	user := f.Locals("authUser").(*models.User)

	file, err := f.FormFile("signature")
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, exception.ParameterError{
			FailedField: "signature",
			Tag:         "required",
			Value:       "",
		})
	}

	signatureURL, err := c.authService().UploadSignature(ctx, user.ID, file)
	if err != nil {
		switch err {
		case exception.ErrRecordNotFound:
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
		case exception.ErrUnsupportedImageType:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.UnsupportedImageTypeResponseError, err)
		case exception.ErrImageFileTooLarge:
			return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.ImageFileTooLargeResponseError, err)
		default:
			return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
		}
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{"signature": signatureURL})
}

// DeleteSignature removes the signed-in user's signature
//
//	@Summary		Delete my signature
//	@Description	Clear the signed-in user's signature and remove the underlying file (best-effort)
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	http_response.OkResponse
//	@Failure		401	{object}	exception.ErrorResponse
//	@Failure		500	{object}	exception.ErrorResponse
//	@Router			/api/v1/auth/profile/signature [delete]
func (c Controller) DeleteSignature(f *fiber.Ctx) error {
	ctx, span := c.m.tracer.TraceStart(f.Context(), "DeleteSignatureController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "DeleteSignature")))

	user := f.Locals("authUser").(*models.User)

	if err := c.authService().DeleteSignature(ctx, user.ID); err != nil {
		if err == exception.ErrRecordNotFound {
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.UnauthorizedResponseError, exception.ErrUnauthorized)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.DbQueryStatementResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return http_response.HttpOkResponse(f, "OK", "Signature deleted successfully")
}
