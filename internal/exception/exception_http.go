package exception

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"maphraohom.app/maphraohom-backoffice/config"
)

func HttpErrorResponseMapping(c *fiber.Ctx, code int, errResponse *ErrorResponse, err error, errParams ...ParameterError) error {
	var (
		isResponseOK        bool
		fiberResponseStatus int
		fiberResponseBody   *ErrorResponse
	)

	// Error mapping
	switch err {
	// Database errors
	case ErrDbQueryStatement:
		errorMessage := &ErrorResponse{
			Code:    DbQueryStatementResponseError.Code,
			Message: fmt.Sprintf("%s, %s", DbQueryStatementResponseError.Message, SqlErrorMessage),
		}
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = errorMessage
	case ErrPromotionOverlap:
		errorMessage := &ErrorResponse{
			Code:    PromotionOverlapResponseError.Code,
			Message: fmt.Sprintf("%s, %s", PromotionOverlapResponseError.Message, PromotionOverlapMessage),
		}
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		fiberResponseBody = errorMessage
	// Application errors
	case ErrInvalidRequestParameter:
		isResponseOK = true
		fiberResponseStatus = fiber.StatusBadRequest
		InvalidRequestParameterResponseError.Errors = errParams
		fiberResponseBody = InvalidRequestParameterResponseError
	default:
		// Default error response
		if errResponse != nil {
			isResponseOK = true
			fiberResponseStatus = code
			fiberResponseBody = errResponse
		} else {
			isResponseOK = false
			fiberResponseStatus = fiber.StatusInternalServerError
			fiberResponseBody = &ErrorResponse{
				Code:    "ER",
				Message: err.Error(),
			}
		}
	}

	// For development
	if !config.IsProduction {
		if isResponseOK {
			if fiberResponseBody != nil {
				return c.Status(fiberResponseStatus).JSON(fiberResponseBody)
			} else {
				return c.SendStatus(fiberResponseStatus)
			}
		}

		// Out of mapping
		return c.Status(fiber.StatusInternalServerError).JSON(&ErrorResponse{Code: "ER", Message: err.Error()})
	}

	// -------------------------------------------------------------------------------------------

	// WARNING! this is for production HTTP response
	// Note: Hide response message for prevent vulnerability attack
	if isResponseOK {
		if fiberResponseBody != nil {
			fiberResponseBody.Message = ""
			return c.Status(fiberResponseStatus).JSON(fiberResponseBody)
		} else {
			return c.SendStatus(fiberResponseStatus)
		}
	}

	// Out of mapping
	return fiber.ErrInternalServerError
}
