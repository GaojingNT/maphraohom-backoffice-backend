package exception

import "github.com/rotisserie/eris"

// ERROR
var (
	// Sql error message
	SqlErrorMessage string
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
	ErrDbQueryStatement = eris.New("database query statement error")
	ErrRecordNotFound   = eris.New("record not found")
	// Bill errors
	ErrBillSequenceMissing = eris.New("bill sequence not seeded for this store and type")
	ErrBillFieldImmutable  = eris.New("storeId and type cannot be changed on an existing bill")
	ErrUnsupportedSlipType = eris.New("slip file type not supported (jpeg, png, webp only)")
	ErrSlipFileTooLarge    = eris.New("slip file exceeds the 10MB size limit")
	// Store errors
	ErrUnsupportedImageType = eris.New("image file type not supported (jpeg, png, webp only)")
	ErrImageFileTooLarge    = eris.New("image file exceeds the 10MB size limit")
	// Auth errors
	ErrInvalidLoginCredential    = eris.New("invalid login credential")
	ErrInvalidResetPasswordToken = eris.New("invalid reset password token")
	ErrEmailAlreadyTaken         = eris.New("email is already used by another user")
	ErrCurrentPasswordIncorrect  = eris.New("current password is incorrect")
	ErrPasswordConfirmMismatch   = eris.New("new password and confirmation do not match")
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
	DbQueryStatementResponseError = &ErrorResponse{Code: "DB-1001", Message: "database query statement error"}
	RecordNotFoundResponseError   = &ErrorResponse{Code: "DB-1002", Message: "record not found"}
	// Bill error response messages
	BillSequenceMissingResponseError = &ErrorResponse{Code: "BILL-1001", Message: "bill sequence not seeded for this store and type"}
	BillFieldImmutableResponseError  = &ErrorResponse{Code: "BILL-1002", Message: "storeId and type cannot be changed on an existing bill"}
	UnsupportedSlipTypeResponseError = &ErrorResponse{Code: "BILL-1003", Message: "slip file type not supported (jpeg, png, webp only)"}
	SlipFileTooLargeResponseError    = &ErrorResponse{Code: "BILL-1004", Message: "slip file exceeds the 10MB size limit"}
	// Store error response messages
	UnsupportedImageTypeResponseError = &ErrorResponse{Code: "STORE-1001", Message: "image file type not supported (jpeg, png, webp only)"}
	ImageFileTooLargeResponseError    = &ErrorResponse{Code: "STORE-1002", Message: "image file exceeds the 10MB size limit"}
	// Auth error response messages
	InvalidLoginCredentialResponseError    = &ErrorResponse{Code: "A-3001", Message: "invalid login credential"}
	InvalidResetPasswordTokenResponseError = &ErrorResponse{Code: "A-3002", Message: "invalid reset password token"}
	EmailAlreadyTakenResponseError         = &ErrorResponse{Code: "A-3003", Message: "email is already used by another user"}
	CurrentPasswordIncorrectResponseError  = &ErrorResponse{Code: "A-3004", Message: "current password is incorrect"}
	PasswordConfirmMismatchResponseError   = &ErrorResponse{Code: "A-3005", Message: "new password and confirmation do not match"}
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
