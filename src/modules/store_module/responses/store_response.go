package responses

import (
	"fmt"

	"github.com/shopspring/decimal"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

// FileURLBuilder turns a stored object key (logo) into a URL the
// client can load directly — the same GET /api/v1/files/* proxy the bill
// module's slip URL uses.
var FileURLBuilder = func(key string) string {
	return fmt.Sprintf("/api/v1/files/%s", key)
}

type (
	// StoreListItem is the shape returned by GET /stores.
	StoreListItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Logo string `json:"logo"`
	}

	// StoreDetailResponse is the shape returned by GET /stores/:id and
	// PUT /stores/:id — every stores table column except deleted_at, with
	// the logo resolved to a ready-to-use URL (empty string when unset), plus
	// the store's owners.
	StoreDetailResponse struct {
		ID        int                  `json:"id"`
		Name      string               `json:"name"`
		Logo      string               `json:"logo"`
		Address   string               `json:"address"`
		Phone     string               `json:"phone"`
		Owners    []StoreOwnerResponse `json:"owners"`
		CreatedAt string               `json:"createdAt"`
		UpdatedAt string               `json:"updatedAt"`
	}

	// StoreOwnerResponse is one of a store's owners (tbl_user_stores).
	StoreOwnerResponse struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
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

// resolveKey turns a stored object key into a ready-to-use URL, leaving an
// unset key as "" rather than a broken "/api/v1/files/" link.
func resolveKey(key string) string {
	if key == "" {
		return ""
	}
	return FileURLBuilder(key)
}

func (StoreListItem) Collection(stores []models.Store) []StoreListItem {
	items := make([]StoreListItem, 0, len(stores))
	for _, store := range stores {
		items = append(items, StoreListItem{
			ID:   store.ID,
			Name: store.Name,
			Logo: resolveKey(store.Logo),
		})
	}
	return items
}

func (response *StoreDetailResponse) Make(store models.Store) *StoreDetailResponse {
	owners := make([]StoreOwnerResponse, 0, len(store.Owners))
	for _, owner := range store.Owners {
		owners = append(owners, StoreOwnerResponse{
			ID:        owner.ID,
			Email:     owner.Email,
			FirstName: owner.FirstName,
			LastName:  owner.LastName,
		})
	}

	return &StoreDetailResponse{
		ID:        store.ID,
		Name:      store.Name,
		Logo:      resolveKey(store.Logo),
		Address:   store.Address,
		Phone:     store.Phone,
		Owners:    owners,
		CreatedAt: store.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: store.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
