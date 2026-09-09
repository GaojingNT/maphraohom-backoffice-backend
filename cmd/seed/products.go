package main

import (
	"log"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// productNames lists all coconut products shared across every store.
var productNames = []string{
	"เนื้อหั่นเส้น",
	"เนื้อหั่นชิ้น",
	"เนื้อเบ้าพับ",
	"เนื้อขูดฝอย",
	"เนื้อริ้ว",
	"น้ำมะพร้าว",
}

// seedProducts inserts all products and returns a name→ID map for price linking.
func seedProducts(db *gorm.DB) map[string]uint {
	productMap := make(map[string]uint, len(productNames))
	for _, name := range productNames {
		p := models.Product{Name: name}
		if err := db.Create(&p).Error; err != nil {
			log.Fatalf("[Seed] product %q: %v", name, err)
		}
		productMap[name] = uint(p.ID)
		log.Printf("[Seed] Product: #%d %s", p.ID, p.Name)
	}
	return productMap
}
