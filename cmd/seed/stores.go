package main

import (
	"log"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

type storeSeed struct {
	Name string

	// BasePrice applies to every product except น้ำมะพร้าว, which is sold
	// by the bottle and priced separately via BottlePrice.
	BasePrice   float64
	BottlePrice float64
}

// storeSeeds defines each store and its per-product base prices.
var storeSeeds = []storeSeed{
	{
		Name:        "มะพร้าวหอมอัมพวา",
		BasePrice:   80,
		BottlePrice: 20,
	},
	{
		Name:        "หอมละมุน มะพร้าวน้ำหอม",
		BasePrice:   100,
		BottlePrice: 25,
	},
}

// seedStores creates each store and its store-product prices. Promotions are
// no longer seeded here — manage those through the admin UI instead.
func seedStores(db *gorm.DB, productMap map[string]uint) {
	for _, ss := range storeSeeds {
		store := models.Store{Name: ss.Name}
		if err := db.Create(&store).Error; err != nil {
			log.Fatalf("[Seed] store %q: %v", ss.Name, err)
		}
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
}
