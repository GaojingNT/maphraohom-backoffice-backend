package responses

import (
	"github.com/shopspring/decimal"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

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

	// LastPriceItem is one product's last price used at a store for a given
	// bill type, returned by GET /stores/:storeId/last-prices — used to
	// prefill the create-bill form (buy and sell prices are never mixed
	// because the query is scoped by type too).
	LastPriceItem struct {
		ProductID int             `json:"productId"`
		Price     decimal.Decimal `json:"price"`
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
