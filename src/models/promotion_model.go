package models

import "time"

// Promotion is a store-wide discount event ("โปรโมชั่น") — one store can
// have at most one active promotion at any given time (enforced by the
// application, not a DB constraint), covering whichever products it sets
// special prices for via PromotionPrice.
type Promotion struct {
	BaseModel

	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	StoreID int    `json:"storeId" gorm:"column:store_id;not null;"`
	Store   *Store `json:"store,omitempty" gorm:"foreignKey:StoreID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	StartsAt time.Time `json:"startsAt" gorm:"column:starts_at;not null;"`
	EndsAt   time.Time `json:"endsAt" gorm:"column:ends_at;not null;"`

	Prices []PromotionPrice `json:"prices,omitempty" gorm:"foreignKey:PromotionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Promotion searchable attributes
func PromotionSearchable() []string {
	return []string{"name"}
}
