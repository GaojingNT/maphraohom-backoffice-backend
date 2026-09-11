package main

import (
	"log"
	"time"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// promotionSeed defines a currently-active test promotion for one store,
// discounting every กก. product (not the bottled น้ำมะพร้าว) to PromoPrice.
type promotionSeed struct {
	StoreKey   string
	Name       string
	PromoPrice float64
}

var promotionSeeds = []promotionSeed{
	{StoreKey: string(storeAmpawa), Name: "โปรราคาพิเศษ มะพร้าวหอมอัมพวา", PromoPrice: 70},
	{StoreKey: string(storeHomlamun), Name: "โปรราคาพิเศษ หอมละมุน", PromoPrice: 90},
}

// seedPromotions creates one active promotion per store for testing the
// "กำลังโปร" price-resolution path end-to-end.
func seedPromotions(db *gorm.DB, storeMap map[string]uint, productMap map[string]uint) {
	now := time.Now()

	for _, ps := range promotionSeeds {
		storeID, ok := storeMap[ps.StoreKey]
		if !ok {
			log.Fatalf("[Seed] store key %q not found", ps.StoreKey)
		}

		promotion := models.Promotion{
			Name:     ps.Name,
			StoreID:  int(storeID),
			StartsAt: now,
			EndsAt:   now.AddDate(0, 0, 30),
		}
		if err := db.Create(&promotion).Error; err != nil {
			log.Fatalf("[Seed] promotion %q: %v", ps.Name, err)
		}
		log.Printf("[Seed] Promotion: #%d %s (store #%d)", promotion.ID, promotion.Name, storeID)

		for _, prod := range productSeeds {
			if prod.Unit == "ขวด" {
				continue
			}

			productID, ok := productMap[prod.Name]
			if !ok {
				log.Fatalf("[Seed] product %q not found", prod.Name)
			}

			promoPrice := models.PromotionPrice{
				PromotionID: promotion.ID,
				ProductID:   int(productID),
				Price:       ps.PromoPrice,
			}
			if err := db.Create(&promoPrice).Error; err != nil {
				log.Fatalf("[Seed] promotion price %q / product %q: %v", ps.Name, prod.Name, err)
			}
		}
		log.Printf("[Seed]   %d promotion prices at %.2f ฿", len(productSeeds)-1, ps.PromoPrice)
	}
}
