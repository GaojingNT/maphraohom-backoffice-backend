package main

import (
	"log"
	"time"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

type productPrice struct {
	Name  string
	Price float64
}

type storeSeed struct {
	Name     string
	Products []productPrice
}

// storeSeeds defines each store and its per-product prices.
var storeSeeds = []storeSeed{
	{
		Name: "มะพร้าวหอมแปรรูป อัมพวา",
		Products: []productPrice{
			{"เนื้อหั่นเส้น", 80},
			{"เนื้อหั่นชิ้น", 80},
			{"เนื้อเบ้าพับ", 80},
			{"เนื้อขูดฝอย", 80},
			{"เนื้อริ้ว", 80},
			{"น้ำมะพร้าว", 50},
		},
	},
	{
		Name: "Maphraohom Ampawa มะพร้าวหอมอัมพวา",
		Products: []productPrice{
			{"เนื้อหั่นเส้น", 100},
			{"เนื้อหั่นชิ้น", 100},
			{"เนื้อเบ้าพับ", 100},
			{"เนื้อขูดฝอย", 100},
			{"เนื้อริ้ว", 100},
			{"น้ำมะพร้าว", 70},
		},
	},
}

// seedStores creates each store and its store-product prices.
func seedStores(db *gorm.DB, productMap map[string]uint) {
	now := time.Now()
	for _, ss := range storeSeeds {
		store := models.Store{Name: ss.Name}
		if err := db.Create(&store).Error; err != nil {
			log.Fatalf("[Seed] store %q: %v", ss.Name, err)
		}
		log.Printf("[Seed] Store: #%d %s", store.ID, store.Name)

		for _, pp := range ss.Products {
			productID, ok := productMap[pp.Name]
			if !ok {
				log.Fatalf("[Seed] product %q not found", pp.Name)
			}
			price := models.StoreProductPrice{
				StoreID:       store.ID,
				ProductID:     int(productID),
				Price:         pp.Price,
				EffectiveFrom: now,
			}
			if err := db.Create(&price).Error; err != nil {
				log.Fatalf("[Seed] price store %q / product %q: %v", ss.Name, pp.Name, err)
			}
			log.Printf("[Seed]   Price: store #%d × product #%d (%s) = %.2f ฿", store.ID, productID, pp.Name, pp.Price)
		}
	}
}
