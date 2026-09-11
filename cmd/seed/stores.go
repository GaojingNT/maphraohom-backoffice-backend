package main

import (
	"log"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// storeKey identifies a seeded store for cross-referencing in promotions.go.
type storeKey string

const (
	storeAmpawa   storeKey = "ampawa"
	storeHomlamun storeKey = "homlamun"
)

type storeSeed struct {
	Key string
	Name string

	// BasePrice applies to every product except น้ำมะพร้าว, which is sold
	// by the bottle and priced separately via BottlePrice.
	BasePrice   float64
	BottlePrice float64
}

// storeSeeds defines each store and its per-product base prices.
var storeSeeds = []storeSeed{
	{
		Key:         string(storeAmpawa),
		Name:        "มะพร้าวหอมอัมพวา",
		BasePrice:   80,
		BottlePrice: 20,
	},
	{
		Key:         string(storeHomlamun),
		Name:        "หอมละมุน มะพร้าวน้ำหอม",
		BasePrice:   100,
		BottlePrice: 25,
	},
}

// seedStores creates each store and its store-product prices, returning a
// store key → ID map for promotion seeding.
func seedStores(db *gorm.DB, productMap map[string]uint) map[string]uint {
	storeMap := make(map[string]uint, len(storeSeeds))

	for _, ss := range storeSeeds {
		store := models.Store{Name: ss.Name}
		if err := db.Create(&store).Error; err != nil {
			log.Fatalf("[Seed] store %q: %v", ss.Name, err)
		}
		storeMap[ss.Key] = uint(store.ID)
		log.Printf("[Seed] Store: #%d %s", store.ID, store.Name)

		for _, ps := range productSeeds {
			productID, ok := productMap[ps.Name]
			if !ok {
				log.Fatalf("[Seed] product %q not found", ps.Name)
			}

			price := ss.BasePrice
			if ps.Unit == "ขวด" {
				price = ss.BottlePrice
			}

			spp := models.StoreProductPrice{
				StoreID:   store.ID,
				ProductID: int(productID),
				Price:     price,
			}
			if err := db.Create(&spp).Error; err != nil {
				log.Fatalf("[Seed] price store %q / product %q: %v", ss.Name, ps.Name, err)
			}
			log.Printf("[Seed]   Price: store #%d × product #%d (%s) = %.2f ฿", store.ID, productID, ps.Name, price)
		}
	}

	return storeMap
}
