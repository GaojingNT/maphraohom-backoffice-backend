package models

type StoreProductPrice struct {
	BaseModel

	// Relations
	StoreID int    `json:"storeId" gorm:"column:store_id;not null;uniqueIndex:idx_store_product_prices_store_product;"`
	Store   *Store `json:"store,omitempty" gorm:"foreignKey:StoreID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	ProductID int      `json:"productId" gorm:"column:product_id;not null;uniqueIndex:idx_store_product_prices_store_product;"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Price is the store's current price for this product. No history is
	// kept — changing the price updates this row in place, and updated_at
	// reflects when it last changed. Past bills are unaffected because
	// BillItem.Price snapshots the price at issue time separately.
	Price float64 `json:"price" gorm:"column:price;not null;"`
}
