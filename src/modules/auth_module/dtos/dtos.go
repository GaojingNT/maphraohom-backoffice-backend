package dtos

type (
	SignInDto struct {
		Username string `json:"username" form:"username" query:"username" validate:"required,email"`
		Password string `json:"password" form:"password" query:"password" validate:"required,min=8"`
	}
	ForgotPasswordDto struct {
		Email string `json:"email" form:"email" query:"email" validate:"required,max=50,email"`
	}
	ResetPasswordDto struct {
		Email              string `json:"email" form:"email" query:"email" validate:"required,max=50,email"`
		Password           string `json:"password" form:"password" query:"password" validate:"required,min=8"`
		ConfirmPassword    string `json:"confirmPassword" form:"confirmPassword" query:"confirmPassword" validate:"required,min=8"`
		ResetPasswordToken string `json:"resetPasswordToken" form:"resetPasswordToken" query:"resetPasswordToken" validate:"required"`
	}
	EncryptPasswordDto struct {
		Password string `json:"password" form:"password" query:"password" validate:"required,min=8"`
	}
)
