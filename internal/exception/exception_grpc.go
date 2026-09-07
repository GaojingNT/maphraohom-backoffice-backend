package exception

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func GrpcErrorResponseMapping(c context.Context, err error) error {
	switch err {
	case ErrDbQueryStatement:
		return ErrDbQueryStatement
	case ErrInvalidToken:
		return fiber.ErrUnauthorized
	case ErrInvalidUserData:
		return ErrInvalidUserData
	case ErrInvalidTokenClaim:
		return ErrInvalidTokenClaim
	case ErrInvalidRequestParameter:
		return ErrInvalidRequestParameter
	case ErrRecordNotFound:
		return ErrRecordNotFound
	}

	// Out of mapping
	return ErrInternalServerError
}
