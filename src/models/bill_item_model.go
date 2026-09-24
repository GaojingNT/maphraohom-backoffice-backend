package models

import "github.com/shopspring/decimal"

type BillItem struct {
	BaseModel

	// Relations
	BillID int `json:"billId" gorm:"column:bill_id;not null;"`

	ProductID int      `json:"productId" gorm:"column:product_id;not null;"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Quantity is in Unit — kilograms ("กก.") for most products, or bottles
	// ("ขวด") for coconut water. Must be a whole number when Unit is "ขวด".
	Quantity decimal.Decimal `json:"quantity" gorm:"column:quantity;type:numeric(10,3);not null;"`

	// Unit is a snapshot of products.unit taken at issue time.
	Unit string `json:"unit" gorm:"column:unit;size:20;not null;"`

	// Price is the per-unit price the user typed in when the bill was
	// created — there is no price list, so this is never resolved from
	// anywhere else.
	Price decimal.Decimal `json:"price" gorm:"column:price;type:numeric(10,2);not null;"`

	// Subtotal = round(quantity * price, 2), computed server-side.
	Subtotal decimal.Decimal `json:"subtotal" gorm:"column:subtotal;type:numeric(12,2);not null;"`
}
