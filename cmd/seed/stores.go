package main

import (
	"log"

	"gorm.io/gorm"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// storeSeeds lists each store. Prices are no longer master data — staff
// type a price per unit into every bill item, so nothing is seeded here.
var storeSeeds = []struct {
	Name string
}{
	{Name: "มะพร้าวหอมอัมพวา"},
	{Name: "หอมละมุน มะพร้าวน้ำหอม"},
}

// billSequenceTypes lists every bill type a fresh store needs a numbering
// row for.
var billSequenceTypes = []string{models.BillTypeReceipt, models.BillTypePayment}

// seedStores creates each store and its bill_sequences rows (one per bill
// type, both starting at book/receipt 0).
func seedStores(db *gorm.DB) {
	for _, ss := range storeSeeds {
		store := models.Store{Name: ss.Name}
		if err := db.Create(&store).Error; err != nil {
			log.Fatalf("[Seed] store %q: %v", ss.Name, err)
		}
		log.Printf("[Seed] Store: #%d %s", store.ID, store.Name)

		for _, billType := range billSequenceTypes {
			sequence := models.BillSequence{StoreID: store.ID, Type: billType}
			if err := db.Create(&sequence).Error; err != nil {
				log.Fatalf("[Seed] bill sequence store %q / type %q: %v", ss.Name, billType, err)
			}
			log.Printf("[Seed]   BillSequence: store #%d × type %s", store.ID, billType)
		}
	}
}
