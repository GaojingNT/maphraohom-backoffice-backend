// Seed initial stores, products, and their prices.
//
// Usage:
//
//	go run cmd/seed/main.go
package main

import (
	"log"
	"time"

	"maphraohom.app/maphraohom-backoffice/config"
	"maphraohom.app/maphraohom-backoffice/pkg/database"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

var storeNames = []string{
	"ร้านสาขา 1",
	"ร้านสาขา 2",
	"ร้านสาขา 3",
}

type productSeed struct {
	Name  string
	Price float64
}

var productSeeds = []productSeed{
	{Name: "มะพร้าวหั่นเส้น 1.5-2ชั้น", Price: 100},
	{Name: "มะพร้าวขูดริ้ว 1.5-2ชั้น", Price: 110},
	{Name: "มะพร้าวขูดฝอย 2ชั้น", Price: 90},
	{Name: "มะพร้าวขูดฝอยทึนทึก", Price: 90},
	{Name: "เนื้อคว้าน 1.5-2ชั้น", Price: 100},
	{Name: "หั่นชิ้น 3 เหลี่ยม 1.5-2ชั้น", Price: 100},
	{Name: "มะพร้าวขูดขลุย (คั้นกะทิ)", Price: 70},
}

func main() {
	config.Global = config.NewConfig()
	dbConn := database.Initialize()

	now := time.Now()

	stores := make([]models.Store, 0, len(storeNames))
	for _, name := range storeNames {
		var store models.Store
		if err := dbConn.Where(models.Store{Name: name}).FirstOrCreate(&store, models.Store{Name: name}).Error; err != nil {
			log.Fatalf("seed store %q: %v", name, err)
		}
		stores = append(stores, store)
		log.Printf("[Seed] Store ready: #%d %s", store.ID, store.Name)
	}

	for _, ps := range productSeeds {
		var product models.Product
		if err := dbConn.Where(models.Product{Name: ps.Name}).FirstOrCreate(&product, models.Product{Name: ps.Name}).Error; err != nil {
			log.Fatalf("seed product %q: %v", ps.Name, err)
		}
		log.Printf("[Seed] Product ready: #%d %s", product.ID, product.Name)

		for _, store := range stores {
			var price models.StoreProductPrice
			err := dbConn.Where(models.StoreProductPrice{StoreID: store.ID, ProductID: product.ID}).
				Attrs(models.StoreProductPrice{Price: ps.Price, EffectiveFrom: now}).
				FirstOrCreate(&price).Error
			if err != nil {
				log.Fatalf("seed price for store #%d / product %q: %v", store.ID, ps.Name, err)
			}
			log.Printf("[Seed]   Price ready: store #%d x product #%d = %.2f", store.ID, product.ID, price.Price)
		}
	}

	log.Println("[Seed] Done.")
}
