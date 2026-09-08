package models

type BillItem struct {
	BaseModel

	// Relations
	BillID int `json:"billId" gorm:"column:bill_id;not null;"`

	ProductID int      `json:"productId" gorm:"column:product_id;not null;"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	Kilogram float64 `json:"kilogram" gorm:"column:kilogram;not null;"`

	// Price snapshot copied from StoreProductPrice (by the parent bill's
	// store_id + this product_id) at issue time.
	Price float64 `json:"price" gorm:"column:price;not null;"`

	// Subtotal = kilogram * price
	Subtotal float64 `json:"subtotal" gorm:"column:subtotal;not null;"`
}
