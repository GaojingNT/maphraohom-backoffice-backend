package responses

import "maphraohom.app/maphraohom-backoffice/src/models"

type (
	// StoreListItem is the shape returned by GET /stores.
	StoreListItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Logo string `json:"logo"`
	}

	// StoreDetailResponse is the shape returned by GET /stores/:id.
	StoreDetailResponse struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Logo      string `json:"logo"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}

	// StoreProductPriceItem is one product's currently effective price at a
	// store, returned by GET /stores/:id/products and .../products/:productId.
	StoreProductPriceItem struct {
		ProductID     int     `json:"productId"`
		ProductName   string  `json:"productName"`
		Price         float64 `json:"price"`
		EffectiveFrom string  `json:"effectiveFrom"`
	}
)

func (StoreListItem) Collection(stores []models.Store) []StoreListItem {
	items := make([]StoreListItem, 0, len(stores))
	for _, store := range stores {
		items = append(items, StoreListItem{
			ID:   store.ID,
			Name: store.Name,
			Logo: store.Logo,
		})
	}
	return items
}

func (response *StoreDetailResponse) Make(store models.Store) *StoreDetailResponse {
	return &StoreDetailResponse{
		ID:        store.ID,
		Name:      store.Name,
		Logo:      store.Logo,
		CreatedAt: store.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: store.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (StoreProductPriceItem) Make(price models.StoreProductPrice) StoreProductPriceItem {
	productName := ""
	if price.Product != nil {
		productName = price.Product.Name
	}

	return StoreProductPriceItem{
		ProductID:     price.ProductID,
		ProductName:   productName,
		Price:         price.Price,
		EffectiveFrom: price.EffectiveFrom.Format("2006-01-02 15:04:05"),
	}
}

func (item StoreProductPriceItem) Collection(prices []models.StoreProductPrice) []StoreProductPriceItem {
	items := make([]StoreProductPriceItem, 0, len(prices))
	for _, price := range prices {
		items = append(items, item.Make(price))
	}
	return items
}
