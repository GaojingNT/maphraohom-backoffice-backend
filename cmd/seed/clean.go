package main

import (
	"log"

	"gorm.io/gorm"
)

// cleanAll truncates all business tables and resets their auto-increment
// sequences. Dependent tables (bill_items, bills, store_product_prices,
// customer_addresses, customer_phones, promotion_prices) are handled
// automatically via CASCADE.
func cleanAll(db *gorm.DB) {
	log.Println("[Clean] Truncating business tables...")
	err := db.Exec(`
		TRUNCATE TABLE
			tbl_bill_items,
			tbl_bills,
			tbl_promotion_prices,
			tbl_promotions,
			tbl_store_product_prices,
			tbl_customer_addresses,
			tbl_customer_phones,
			tbl_customers,
			tbl_products,
			tbl_stores
		RESTART IDENTITY CASCADE
	`).Error
	if err != nil {
		log.Fatalf("[Clean] truncate failed: %v", err)
	}
	log.Println("[Clean] Done.")
}
