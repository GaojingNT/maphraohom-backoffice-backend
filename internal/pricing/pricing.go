// Package pricing resolves the price a store charges for a product at a
// given time, preferring an active promotion price over the store's base
// price. It is shared by bill_module (pricing bill items) and store_module
// (previewing resolved prices before a bill is created).
package pricing

import (
	"time"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// Resolved is one product's resolved price.
type Resolved struct {
	Price       float64
	IsPromotion bool
	PromotionID *int
}

// ResolveOne resolves a single product's price at a store at time `at`.
// Returns gorm.ErrRecordNotFound if neither a promotion nor a base price
// exists for this store+product.
func ResolveOne(tx *gorm.DB, storeID int, productID int, at time.Time) (Resolved, error) {
	var promo struct {
		PromotionID int
		Price       float64
	}
	if err := tx.Model(&models.PromotionPrice{}).
		Select("tbl_promotions.id AS promotion_id, tbl_promotion_prices.price AS price").
		Joins("JOIN tbl_promotions ON tbl_promotions.id = tbl_promotion_prices.promotion_id").
		Where("tbl_promotions.store_id = ? AND tbl_promotion_prices.product_id = ? AND ? BETWEEN tbl_promotions.starts_at AND tbl_promotions.ends_at", storeID, productID, at).
		Order("tbl_promotions.starts_at DESC").
		Limit(1).
		Scan(&promo).Error; err != nil {
		return Resolved{}, err
	}
	if promo.PromotionID != 0 {
		id := promo.PromotionID
		return Resolved{Price: promo.Price, IsPromotion: true, PromotionID: &id}, nil
	}

	var price models.StoreProductPrice
	if err := tx.Where("store_id = ? AND product_id = ?", storeID, productID).First(&price).Error; err != nil {
		return Resolved{}, err
	}
	return Resolved{Price: price.Price, IsPromotion: false, PromotionID: nil}, nil
}

// ActivePromotion returns the store's currently active promotion (if any)
// with its per-product prices preloaded. Returns gorm.ErrRecordNotFound if
// none is active. A store can have at most one active promotion at a time
// (enforced when promotions are created), so there is never more than one to
// pick among.
func ActivePromotion(tx *gorm.DB, storeID int, at time.Time) (models.Promotion, error) {
	var promotion models.Promotion
	err := tx.Preload("Prices").
		Where("store_id = ? AND starts_at <= ? AND ends_at >= ?", storeID, at, at).
		Order("starts_at DESC").
		First(&promotion).Error
	return promotion, err
}
