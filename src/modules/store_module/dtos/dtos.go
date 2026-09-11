package dtos

// UpdateStoreProductPrice sets a store's base price for one product.
// store_product_prices keeps no history — this updates the single row for
// this (store, product) pair in place (or creates it if it doesn't exist
// yet).
type UpdateStoreProductPrice struct {
	Price float64 `json:"price" validate:"required,gt=0"`
}
