package dtos

type CreateBill struct {
	StoreID         int     `form:"storeId" validate:"required"`
	CustomerName    string  `form:"customerName" validate:"required"`
	CustomerAddress string  `form:"customerAddress" validate:"required"`
	Discount        float64 `form:"discount"`
	ShippingFee     float64 `form:"shippingFee"`

	// Items is a JSON-encoded array of CreateBillItem, e.g.:
	// [{"productId":1,"kilogram":2.5},{"productId":3,"kilogram":1.2}]
	// (kept as a plain form field since multipart/form-data has no native
	// array-of-objects encoding).
	Items string `form:"items" validate:"required"`
}

type CreateBillItem struct {
	ProductID int     `json:"productId" validate:"required"`
	Kilogram  float64 `json:"kilogram" validate:"required,gt=0"`
}

// UpdateBill replaces a bill's editable fields and its full set of line
// items (existing items are deleted and recreated, not diffed). Book/receipt
// numbers never change on edit. RemoveSlip clears an existing slip when no
// new one is uploaded in its place.
type UpdateBill struct {
	StoreID         int     `form:"storeId" validate:"required"`
	CustomerName    string  `form:"customerName" validate:"required"`
	CustomerAddress string  `form:"customerAddress" validate:"required"`
	Discount        float64 `form:"discount"`
	ShippingFee     float64 `form:"shippingFee"`
	Items           string  `form:"items" validate:"required"`
	RemoveSlip      bool    `form:"removeSlip"`
}
