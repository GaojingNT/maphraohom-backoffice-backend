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

	// StoreProductPriceItem is one product's currently resolved price at a
	// store (promotion price if one is active, else the base price),
	// returned by GET /stores/:id/products and .../products/:productId.
	StoreProductPriceItem struct {
		ProductID   int     `json:"productId"`
		ProductName string  `json:"productName"`
		Unit        string  `json:"unit"`
		Price       float64 `json:"price"`
		IsPromotion bool    `json:"isPromotion"`
		PromotionID *int    `json:"promotionId,omitempty"`
	}

	// StoreProductBasePriceItem is one product's editable base price at a
	// store — the raw store_product_prices row, never resolved against an
	// active promotion. Used by the price-management admin screen, where
	// showing/editing a promo-discounted number would be wrong.
	StoreProductBasePriceItem struct {
		ProductID   int     `json:"productId"`
		ProductName string  `json:"productName"`
		Unit        string  `json:"unit"`
		Price       float64 `json:"price"`
		UpdatedAt   string  `json:"updatedAt"`
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

// Make builds one product's resolved price. When promotion is non-nil and
// carries a special price for this product, that price wins; otherwise the
// store's base price is used.
func (StoreProductPriceItem) Make(price models.StoreProductPrice, promotion *models.Promotion) StoreProductPriceItem {
	productName, unit := "", ""
	if price.Product != nil {
		productName = price.Product.Name
		unit = price.Product.Unit
	}

	item := StoreProductPriceItem{
		ProductID:   price.ProductID,
		ProductName: productName,
		Unit:        unit,
		Price:       price.Price,
	}

	if promotion != nil {
		for _, promoPrice := range promotion.Prices {
			if promoPrice.ProductID == price.ProductID {
				promotionID := promotion.ID
				item.Price = promoPrice.Price
				item.IsPromotion = true
				item.PromotionID = &promotionID
				break
			}
		}
	}

	return item
}

func (item StoreProductPriceItem) Collection(prices []models.StoreProductPrice, promotion *models.Promotion) []StoreProductPriceItem {
	items := make([]StoreProductPriceItem, 0, len(prices))
	for _, price := range prices {
		items = append(items, item.Make(price, promotion))
	}
	return items
}

func (StoreProductBasePriceItem) Make(price models.StoreProductPrice) StoreProductBasePriceItem {
	productName, unit := "", ""
	if price.Product != nil {
		productName = price.Product.Name
		unit = price.Product.Unit
	}

	return StoreProductBasePriceItem{
		ProductID:   price.ProductID,
		ProductName: productName,
		Unit:        unit,
		Price:       price.Price,
		UpdatedAt:   price.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (item StoreProductBasePriceItem) Collection(prices []models.StoreProductPrice) []StoreProductBasePriceItem {
	items := make([]StoreProductBasePriceItem, 0, len(prices))
	for _, price := range prices {
		items = append(items, item.Make(price))
	}
	return items
}
