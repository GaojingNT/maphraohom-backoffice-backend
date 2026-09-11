package exception

import "github.com/rotisserie/eris"

// ERROR
var (
	// Sql error message
	SqlErrorMessage string
	// Promotion overlap detail message (names the conflicting promotion)
	PromotionOverlapMessage string
	// Server errors
	ErrInternalServerError            = eris.New("internal server error")
	ErrUnauthorized                   = eris.New("unauthorized")
	ErrForbidden                      = eris.New("forbidden")
	ErrNotFound                       = eris.New("not found")
	ErrTypeConversionFailed           = eris.New("type conversion failed")
	ErrInvalidRequestBody             = eris.New("invalid request body")
	ErrFileTypeNotSupported           = eris.New("file type not supported")
	ErrFileSizeLimitExceeded          = eris.New("file size limit exceeded")
	ErrRequestContentTypeNotSupported = eris.New("request content type not supported")
	// Database errors
	ErrDbQueryStatement   = eris.New("database query statement error")
	ErrRecordNotFound     = eris.New("record not found")
	ErrPriceNotConfigured = eris.New("price not configured for this store and product")
	ErrPromotionOverlap   = eris.New("promotion overlaps with an existing active promotion for this store")
	// Auth errors
	ErrInvalidLoginCredential    = eris.New("invalid login credential")
	ErrInvalidResetPasswordToken = eris.New("invalid reset password token")
	// token errors
	ErrInvalidToken            = eris.New("invalid token")
	ErrInvalidRequestParameter = eris.New("invalid request parameter")
	ErrInvalidUserData         = eris.New("invalid user data")
	ErrInvalidTokenClaim       = eris.New("invalid token claim")
	ErrInvalidTokenClaimData   = eris.New("invalid token claim data")
	ErrTokenExpired            = eris.New("token expired")
)

// MESSAGE RESPONSE
var (
	// Server error response messages
	NotFoundResponseError            = &ErrorResponse{Code: "S-404", Message: "not found"}
	ErrorResponseInternalServerError = &ErrorResponse{Code: "S-500", Message: "internal server error"}
	// Database error response messages
	DbQueryStatementResponseError   = &ErrorResponse{Code: "DB-1001", Message: "database query statement error"}
	RecordNotFoundResponseError     = &ErrorResponse{Code: "DB-1002", Message: "record not found"}
	PriceNotConfiguredResponseError = &ErrorResponse{Code: "DB-1003", Message: "price not configured for this store and product"}
	PromotionOverlapResponseError   = &ErrorResponse{Code: "DB-1004", Message: "promotion overlaps with an existing active promotion for this store"}
	// Auth error response messages
	InvalidLoginCredentialResponseError    = &ErrorResponse{Code: "A-3001", Message: "invalid login credential"}
	InvalidResetPasswordTokenResponseError = &ErrorResponse{Code: "A-3002", Message: "invalid reset password token"}
	// token error response messages
	UnauthorizedResponseError            = &ErrorResponse{Code: "T-1001", Message: "unauthorized"}
	ForbiddenResponseError               = &ErrorResponse{Code: "T-1002", Message: "forbidden"}
	InvalidTokenResponseError            = &ErrorResponse{Code: "T-2003", Message: "invalid token"}
	InvalidRequestParameterResponseError = &ErrorResponse{Code: "T-2004", Message: "invalid request parameter"}
	InvalidUserDataResponseError         = &ErrorResponse{Code: "T-2005", Message: "invalid user data"}
	InvalidTokenClaimResponseError       = &ErrorResponse{Code: "T-2006", Message: "invalid token claim"}
	InvalidTokenClaimDataResponseError   = &ErrorResponse{Code: "T-2006", Message: "invalid token claim data"}
	TokenExpiredResponseError            = &ErrorResponse{Code: "T-2007", Message: "token expired"}
)
