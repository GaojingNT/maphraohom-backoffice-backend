package main

import (
	"log"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// productSeeds lists all coconut products shared across every store, in
// their selling unit. All but น้ำมะพร้าว are sold by the kilogram; น้ำมะพร้าว
// is sold by the bottle.
var productSeeds = []struct {
	Name string
	Unit string
}{
	{"เนื้อพับ 1.5 ชั้น", "กก."},
	{"เนื้อ 2 ชั้น", "กก."},
	{"เนื้อพุดดิ้ง", "กก."},
	{"เนื้อริ้ว 1.2 ชั้น", "กก."},
	{"เนื้อริ้ว 1.5 ชั้น", "กก."},
	{"เนื้อริ้ว 1.7-2 ชั้น", "กก."},
	{"เนื้อหั่น 1.5 ชั้น", "กก."},
	{"เนื้อหั่น 2 ชั้น", "กก."},
	{"เนื้อหั่นเส้น 5 มิล", "กก."},
	{"เนื้อหั่นเส้น 10 มิล", "กก."},
	{"ฝอยเนื้อ 2 ชั้น", "กก."},
	{"ฝอยเนื้อทึนทึก", "กก."},
	{"เนื้อขูดละเอียด", "กก."},
	{"น้ำมะพร้าว", "ขวด"},
}

// seedProducts inserts all products and returns a name→ID map for price linking.
func seedProducts(db *gorm.DB) map[string]uint {
	productMap := make(map[string]uint, len(productSeeds))
	for _, ps := range productSeeds {
		p := models.Product{Name: ps.Name, Unit: ps.Unit}
		if err := db.Create(&p).Error; err != nil {
			log.Fatalf("[Seed] product %q: %v", ps.Name, err)
		}
		productMap[ps.Name] = uint(p.ID)
		log.Printf("[Seed] Product: #%d %s (%s)", p.ID, p.Name, p.Unit)
	}
	return productMap
}
