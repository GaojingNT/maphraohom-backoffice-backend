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
	// UpdateProfileDto is the JSON body of PUT /api/v1/auth/profile. The
	// signature is uploaded through its own endpoint
	// (PUT /api/v1/auth/profile/signature).
	UpdateProfileDto struct {
		FirstName string `json:"firstName" form:"firstName" validate:"max=100"`
		LastName  string `json:"lastName" form:"lastName" validate:"max=100"`
		Email     string `json:"email" form:"email" validate:"required,max=100,email"`
	}
)
