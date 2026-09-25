package main

import (
	"log"

	"gorm.io/gorm"
)

// protectedTables hold real business records the seed must never wipe.
var protectedTables = []string{
	"tbl_bills",
	"tbl_bill_items",
	"tbl_customers",
	"tbl_customer_addresses",
	"tbl_customer_phones",
}

// ensureNoBusinessData aborts the seed if any protected table has a row
// (soft-deleted rows count — they're still real history). cleanAll's
// TRUNCATE ... CASCADE would otherwise take them with it.
func ensureNoBusinessData(db *gorm.DB) {
	for _, table := range protectedTables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			log.Fatalf("[Seed] count %s: %v", table, err)
		}
		if count > 0 {
			log.Fatalf("[Seed] REFUSING to seed: %s has %d row(s). The full seed truncates stores/products, "+
				"which would cascade into bills and customers. Use `go run ./cmd/seed/ owner` to only add a "+
				"login user, or back up and clear the data by hand if you really mean to start over.", table, count)
		}
	}
}

// cleanAll truncates all business tables and resets their auto-increment
// sequences. Only ever called after ensureNoBusinessData, so the bill and
// customer tables listed here are already empty — they're included just so
// the identity sequences restart at 1. Dependent tables (bill_sequences,
// user_stores) are handled automatically via CASCADE.
func cleanAll(db *gorm.DB) {
	log.Println("[Clean] Truncating business tables...")
	err := db.Exec(`
		TRUNCATE TABLE
			tbl_bill_items,
			tbl_bills,
			tbl_bill_sequences,
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
