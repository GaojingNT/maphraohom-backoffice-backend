package models

import "time"

type StoreProductPrice struct {
	BaseModel

	// Relations
	StoreID int    `json:"storeId" gorm:"column:store_id;not null;"`
	Store   *Store `json:"store,omitempty" gorm:"foreignKey:StoreID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	ProductID int      `json:"productId" gorm:"column:product_id;not null;"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Fields
	Price         float64   `json:"price" gorm:"column:price;not null;"`
	EffectiveFrom time.Time `json:"effectiveFrom" gorm:"column:effective_from;not null;"`
}
