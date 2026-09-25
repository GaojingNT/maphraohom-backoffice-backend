// Seed initial stores, products, and a login user who owns every store.
//
// Usage:
//
//	go run ./cmd/seed/         # fresh database only — see below
//	go run ./cmd/seed/ owner   # just create/refresh the owner login; touches nothing else
//
// The full seed truncates stores/products before re-creating them, so it
// refuses to run while any bill or customer row exists (soft-deleted ones
// included) — bill and customer data must never be wiped by a seed. On a
// database that already has real data, use `owner` mode instead.
package main

import (
	"os"

	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
)

func main() {
	config.Global = config.NewConfig()
	db := database.Initialize()

	if len(os.Args) > 1 && os.Args[1] == "owner" {
		seedOwner(db)
		return
	}

	ensureNoBusinessData(db)
	cleanAll(db)

	seedProducts(db)
	seedStores(db)
	seedOwner(db)
}
