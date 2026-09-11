package models

// PromotionPrice is one product's special price within a Promotion.
type PromotionPrice struct {
	BaseModel

	PromotionID int        `json:"promotionId" gorm:"column:promotion_id;not null;"`
	Promotion   *Promotion `json:"promotion,omitempty" gorm:"foreignKey:PromotionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	ProductID int      `json:"productId" gorm:"column:product_id;not null;"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	Price float64 `json:"price" gorm:"column:price;not null;"`
}
