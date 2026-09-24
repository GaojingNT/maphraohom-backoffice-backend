package dtos

type CreateCustomer struct {
	Name string `json:"name" validate:"required"`
}

type CreateCustomerAddress struct {
	Address   string `json:"address" validate:"required"`
	IsDefault bool   `json:"isDefault"`
}

type CreateCustomerPhone struct {
	Phone     string `json:"phone" validate:"required"`
	IsDefault bool   `json:"isDefault"`
}
