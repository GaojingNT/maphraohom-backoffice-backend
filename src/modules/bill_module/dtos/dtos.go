package dtos

type CreateBill struct {
	ProductID       int     `form:"productId" validate:"required"`
	StoreID         int     `form:"storeId" validate:"required"`
	CustomerName    string  `form:"customerName" validate:"required"`
	CustomerAddress string  `form:"customerAddress" validate:"required"`
	Kilogram        float64 `form:"kilogram" validate:"required,gt=0"`
	Discount        float64 `form:"discount"`
	ShippingFee     float64 `form:"shippingFee"`
}
