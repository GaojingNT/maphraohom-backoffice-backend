package main

import (
	"log"
	"os"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/internal/encryption"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// defaultOwnerEmail / defaultOwnerPassword are used when SEED_OWNER_EMAIL /
// SEED_OWNER_PASSWORD aren't set — local development only. Set both env vars
// when seeding anything reachable from outside.
const (
	defaultOwnerEmail    = "owner@maphraohom.local"
	defaultOwnerPassword = "maphraohom1234"
)

// seedOwner creates (or reuses, by email) one login user and makes them an
// owner of every seeded store. Unlike the business tables, tbl_users is NOT
// truncated by cleanAll — an existing user keeps their password, name and
// signature; only their store ownership is re-attached (tbl_user_stores is
// emptied by the tbl_stores TRUNCATE ... CASCADE).
func seedOwner(db *gorm.DB) {
	email := os.Getenv("SEED_OWNER_EMAIL")
	if email == "" {
		email = defaultOwnerEmail
	}
	password := os.Getenv("SEED_OWNER_PASSWORD")
	if config.IsProduction && (os.Getenv("SEED_OWNER_EMAIL") == "" || len(password) < 8) {
		log.Fatal("[Seed] In production, set SEED_OWNER_EMAIL and SEED_OWNER_PASSWORD (8+ characters) — the default dev login is never used there.")
	}
	if password == "" {
		password = defaultOwnerPassword
		log.Printf("[Seed] SEED_OWNER_PASSWORD not set — using the default dev password for %s", email)
	}

	var user models.User
	err := db.Where("email = ?", email).
		Attrs(models.User{
			Email:     email,
			FirstName: "เจ้าของ",
			LastName:  "ร้าน",
			Password:  encryption.EncryptPassword(password, ""),
		}).
		FirstOrCreate(&user).Error
	if err != nil {
		log.Fatalf("[Seed] owner %q: %v", email, err)
	}
	log.Printf("[Seed] Owner: #%d %s", user.ID, user.Email)

	var stores []models.Store
	if err := db.Find(&stores).Error; err != nil {
		log.Fatalf("[Seed] list stores: %v", err)
	}
	if err := db.Model(&user).Association("Stores").Replace(stores); err != nil {
		log.Fatalf("[Seed] owner %q stores: %v", email, err)
	}
	for _, store := range stores {
		log.Printf("[Seed]   Owns: store #%d %s", store.ID, store.Name)
	}
}
