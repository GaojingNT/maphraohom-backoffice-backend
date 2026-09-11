package responses

import "maphraohom.app/maphraohom-backoffice/src/models"

// ProductListItem is the shape returned by GET /products.
type ProductListItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

func (ProductListItem) Collection(products []models.Product) []ProductListItem {
	items := make([]ProductListItem, 0, len(products))
	for _, product := range products {
		items = append(items, ProductListItem{
			ID:   product.ID,
			Name: product.Name,
			Unit: product.Unit,
		})
	}
	return items
}
