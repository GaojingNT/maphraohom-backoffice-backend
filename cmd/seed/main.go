// Seed initial stores, products, and their prices.
// WARNING: This will truncate all business data before seeding.
//
// Usage:
//
//	go run ./cmd/seed/
package main

import (
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
)

func main() {
	config.Global = config.NewConfig()
	db := database.Initialize()

	cleanAll(db)

	productMap := seedProducts(db)
	seedStores(db, productMap)
}
