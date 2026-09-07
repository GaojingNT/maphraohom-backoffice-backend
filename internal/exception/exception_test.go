package exception

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestHttpErrorResponseMapping(t *testing.T) {
	type args struct {
		c         *fiber.Ctx
		code      int
		message   *ErrorResponse
		err       error
		errParams []ParameterError
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Database query statement error",
			args: args{
				c:    fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code: fiber.StatusBadRequest,
				message: &ErrorResponse{
					Code:    DbQueryStatementResponseError.Code,
					Message: DbQueryStatementResponseError.Message,
				},
				err: ErrDbQueryStatement,
			},
			wantErr: false,
		},
		{
			name: "Invalid request parameter error",
			args: args{
				c:    fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code: fiber.StatusBadRequest,
				message: &ErrorResponse{
					Code:    InvalidRequestParameterResponseError.Code,
					Message: InvalidRequestParameterResponseError.Message,
				},
				err: ErrInvalidRequestParameter,
				errParams: []ParameterError{
					{FailedField: "field1", Tag: "required", Value: "value1"},
				},
			},
			wantErr: false,
		},
		{
			name: "Unauthorized error",
			args: args{
				c:       fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code:    fiber.StatusUnauthorized,
				message: &ErrorResponse{
					Code:    UnauthorizedResponseError.Code,
					Message: UnauthorizedResponseError.Message,
				},
				err:     ErrUnauthorized,
			},
			wantErr: false,
		},
		{
			name: "Forbidden error",
			args: args{
				c:       fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code:    fiber.StatusForbidden,
				message: &ErrorResponse{
					Code:    ForbiddenResponseError.Code,
					Message: ForbiddenResponseError.Message,
				},
				err:     ErrForbidden,
			},
			wantErr: false,
		},
		{
			name: "Record not found error",
			args: args{
				c:       fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code:    fiber.StatusNotFound,
				message: &ErrorResponse{
					Code:    NotFoundResponseError.Code,
					Message: NotFoundResponseError.Message,
				},
				err:     ErrRecordNotFound,
			},
			wantErr: false,
		},
		{
			name: "Invalid token error",
			args: args{
				c:       fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code:    fiber.StatusUnauthorized,
				message: &ErrorResponse{
					Code:    InvalidTokenResponseError.Code,
					Message: InvalidTokenResponseError.Message,
				},
				err:     ErrInvalidToken,
			},
			wantErr: false,
		},
		{
			name: "Invalid user data error",
			args: args{
				c:    fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code: fiber.StatusBadRequest,
				message: &ErrorResponse{
					Code:    DbQueryStatementResponseError.Code,
					Message: DbQueryStatementResponseError.Message,
				},
				err: ErrInvalidUserData,
			},
			wantErr: false,
		},
		{
			name: "Invalid token claim error",
			args: args{
				c:       fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code:    fiber.StatusUnauthorized,
				message: &ErrorResponse{
					Code:    InvalidTokenClaimResponseError.Code,
					Message: InvalidTokenClaimResponseError.Message,
				},
				err:     ErrInvalidTokenClaim,
			},
			wantErr: false,
		},
		{
			name: "Token expired error",
			args: args{
				c:       fiber.New().AcquireCtx(&fasthttp.RequestCtx{}),
				code:    fiber.StatusUnauthorized,
				message: &ErrorResponse{
					Code:    TokenExpiredResponseError.Code,
					Message: TokenExpiredResponseError.Message,
				},
				err:     ErrTokenExpired,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HttpErrorResponseMapping(tt.args.c, tt.args.code, tt.args.message, tt.args.err, tt.args.errParams...); (err != nil) != tt.wantErr {
				t.Errorf("HttpErrorResponseMapping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGrpcErrorResponseMapping(t *testing.T) {
	type args struct {
		c   context.Context
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Database query statement error",
			args: args{
				c:   context.Background(),
				err: ErrDbQueryStatement,
			},
			wantErr: true,
		},
		{
			name: "Invalid request parameter error",
			args: args{
				c:   context.Background(),
				err: ErrInvalidRequestParameter,
			},
			wantErr: true,
		},
		{
			name: "Unauthorized error",
			args: args{
				c:   context.Background(),
				err: ErrUnauthorized,
			},
			wantErr: true,
		},
		{
			name: "Forbidden error",
			args: args{
				c:   context.Background(),
				err: ErrForbidden,
			},
			wantErr: true,
		},
		{
			name: "Record not found error",
			args: args{
				c:   context.Background(),
				err: ErrRecordNotFound,
			},
			wantErr: true,
		},
		{
			name: "Invalid token error",
			args: args{
				c:   context.Background(),
				err: ErrInvalidToken,
			},
			wantErr: true,
		},
		{
			name: "Invalid user data error",
			args: args{
				c:   context.Background(),
				err: ErrInvalidUserData,
			},
			wantErr: true,
		},
		{
			name: "Invalid token claim error",
			args: args{
				c:   context.Background(),
				err: ErrInvalidTokenClaim,
			},
			wantErr: true,
		},
		{
			name: "Token expired error",
			args: args{
				c:   context.Background(),
				err: ErrTokenExpired,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GrpcErrorResponseMapping(tt.args.c, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("GrpcErrorResponseMapping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
