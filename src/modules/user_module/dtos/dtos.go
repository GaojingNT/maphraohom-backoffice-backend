package dtos

type (
	CreateUser struct {
		FirstName string `json:"firstName" form:"firstName" query:"firstName" validate:"required,max:50"`
		LastName  string `json:"lastName" form:"lastName" query:"lastName" validate:"required,max:50"`
		Email     string `json:"email" form:"email" query:"email" validate:"required,email,max:100"`
		RoleID    *int   `json:"roleId" form:"roleId" query:"roleId"`
	}
	UpdateUser struct {
		FirstName string `json:"firstName" form:"firstName" query:"firstName" validate:"required,max:50"`
		LastName  string `json:"lastName" form:"lastName" query:"lastName" validate:"required,max:50"`
		Email     string `json:"email" form:"email" query:"email" validate:"required,email,max:100"`
		RoleID    *int   `json:"roleId" form:"roleId" query:"roleId"`
	}
)
