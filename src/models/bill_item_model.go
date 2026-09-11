package models

type BillItem struct {
	BaseModel

	// Relations
	BillID int `json:"billId" gorm:"column:bill_id;not null;"`

	ProductID int      `json:"productId" gorm:"column:product_id;not null;"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// PromotionID is set when this item's price came from an active
	// promotion instead of the store's base price — nil otherwise.
	PromotionID *int       `json:"promotionId,omitempty" gorm:"column:promotion_id;null;"`
	Promotion   *Promotion `json:"promotion,omitempty" gorm:"foreignKey:PromotionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Quantity is in the product's unit — kilograms for most products, or
	// bottles ("ขวด") for coconut water.
	Quantity float64 `json:"quantity" gorm:"column:quantity;not null;"`

	// Price snapshot resolved (promotion-first, else store base price) at
	// issue time.
	Price float64 `json:"price" gorm:"column:price;not null;"`

	// Subtotal = quantity * price
	Subtotal float64 `json:"subtotal" gorm:"column:subtotal;not null;"`
}
