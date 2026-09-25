package auth_module

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/encryption"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
	"maphraohom.app/maphraohom-backoffice/internal/validator"
	"maphraohom.app/maphraohom-backoffice/src/models"
	"maphraohom.app/maphraohom-backoffice/src/modules/auth_module/dtos"
)

// SignIn sign in
//
//	@Summary		Sign in
//	@Description	Sign in
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			username	body		string	true	"username"
//	@Param			password	body		string	true	"password"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		400			{object}	http_response.OkResponse
//	@Failure		500			{object}	http_response.OkResponse
//	@Router			/api/v1/auth/signin [post]
func (c Controller) SignIn(f *fiber.Ctx) error {
	var (
		ctx, span   = c.m.tracer.TraceStart(f.Context(), "SignInController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "SignIn")))
		accessToken string
		err         error
	)

	// Create data transfer object
	dto := new(dtos.SignInDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	accessToken, err = c.authService().Authenticate(ctx, dto)
	if err != nil {
		if err == exception.ErrInvalidLoginCredential {
			return exception.HttpErrorResponseMapping(f, fiber.StatusUnauthorized, exception.InvalidLoginCredentialResponseError, err)
		}
		return exception.HttpErrorResponseMapping(f, fiber.StatusInternalServerError, exception.ErrorResponseInternalServerError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":        0,
		"message":     "OK",
		"accessToken": accessToken,
		"tokenType":   "Bearer",
		"expiresIn":   config.Global.Auth.SessionLifetime,
	})
}

// ForgotPassword request forgot password
//
//	@Summary		Forgot password
//	@Description	Request to forgot password
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			email	body		string	true	"email"
//	@Success		200		{object}	http_response.OkResponse
//	@Failure		400		{object}	http_response.OkResponse
//	@Failure		500		{object}	http_response.OkResponse
//	@Router			/api/v1/auth/forgot-password [post]
func (c Controller) ForgotPassword(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ForgotPasswordController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ForgotPassword")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.ForgotPasswordDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.authService().ForgotPassword(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// ResetPassword request reset password
//
//	@Summary		reset password
//	@Description	Request to reset password
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			email				body		string	true	"email"
//	@Param			resetPasswordToken	body		string	true	"resetPasswordToken"
//	@Param			password			body		string	true	"password"
//	@Param			confirmPassword		body		string	true	"confirmPassword"
//	@Success		200					{object}	http_response.OkResponse
//	@Failure		400					{object}	http_response.OkResponse
//	@Failure		500					{object}	http_response.OkResponse
//	@Router			/api/v1/auth/reset-password [post]
func (c Controller) ResetPassword(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "ResetPasswordController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "ResetPassword")))
		err       error
	)

	// Create data transfer object
	dto := new(dtos.ResetPasswordDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	err = c.authService().ResetPassword(ctx, dto)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    0,
		"message": "OK",
	})
}

// EncryptPassword encrypt password
//
//	@Summary		Encrypt password
//	@Description	Encrypt password
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Param			password	body		string	true	"password"
//	@Success		200			{object}	http_response.OkResponse
//	@Failure		400			{object}	http_response.OkResponse
//	@Failure		500			{object}	http_response.OkResponse
//	@Router			/api/v1/auth/encrypt-password [post]
func (c Controller) EncryptPassword(f *fiber.Ctx) error {
	var (
		_, span = c.m.tracer.TraceStart(f.Context(), "EncryptPasswordController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "EncryptPassword")))
	)

	// Create data transfer object
	dto := new(dtos.EncryptPasswordDto)

	// Parse HTTP request body to struct variable
	if err := f.BodyParser(dto); err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	// Form request validation
	errors := validator.Validate(*dto)
	if errors != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, exception.ErrInvalidRequestParameter, errors...)
	}

	// Call service function
	encryptedPassword := encryption.EncryptPassword(dto.Password, "")

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":              0,
		"message":           "OK",
		"encryptedPassword": encryptedPassword,
	})
}

// GetProfile get profile
//
//	@Summary		Get profile
//	@Description	Get profile
//	@Tags			Auth Module (Version 1)
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.GetProfileResponse
//	@Failure		400	{object}	http_response.OkResponse
//	@Failure		401	{object}	http_response.OkResponse
//	@Failure		500	{object}	http_response.OkResponse
//	@Router			/api/v1/auth/profile [get]
func (c Controller) GetProfile(f *fiber.Ctx) error {
	var (
		ctx, span = c.m.tracer.TraceStart(f.Context(), "GetProfileController", trace.WithAttributes(attribute.String("server", "http"), attribute.String("controller", "GetProfile")))
		err       error
	)

	// Get user from context
	user := f.Locals("authUser").(*models.User)

	// Call service function
	response, err := c.authService().GetProfile(ctx, user.ID)
	if err != nil {
		return exception.HttpErrorResponseMapping(f, fiber.StatusBadRequest, exception.InvalidRequestParameterResponseError, err)
	}

	c.m.tracer.TraceEnd(span)
	return f.Status(fiber.StatusOK).JSON(response)
}
