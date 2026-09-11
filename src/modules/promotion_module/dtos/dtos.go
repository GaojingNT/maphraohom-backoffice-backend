package dtos

import "time"

// CreatePromotion creates a store-wide promotion with special prices for one
// or more products. A store may have at most one active promotion for any
// given time range — creation is rejected if it overlaps an existing one.
type CreatePromotion struct {
	Name     string                `json:"name" validate:"required"`
	StoreID  int                   `json:"storeId" validate:"required"`
	StartsAt time.Time             `json:"startsAt" validate:"required"`
	EndsAt   time.Time             `json:"endsAt" validate:"required,gtfield=StartsAt"`
	Items    []CreatePromotionItem `json:"items" validate:"required,min=1"`
}

type CreatePromotionItem struct {
	ProductID int     `json:"productId" validate:"required"`
	Price     float64 `json:"price" validate:"required,gt=0"`
}

// UpdatePromotion replaces a promotion's editable fields and its full set of
// per-product prices (existing prices are deleted and recreated, not
// diffed) — same overlap rule as create, excluding the promotion itself.
type UpdatePromotion struct {
	Name     string                `json:"name" validate:"required"`
	StoreID  int                   `json:"storeId" validate:"required"`
	StartsAt time.Time             `json:"startsAt" validate:"required"`
	EndsAt   time.Time             `json:"endsAt" validate:"required,gtfield=StartsAt"`
	Items    []CreatePromotionItem `json:"items" validate:"required,min=1"`
}
