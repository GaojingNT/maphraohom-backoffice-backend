package responses

import (
	"time"

	"maphraohom.app/maphraohom-backoffice/src/models"
)

type (
	// PromotionListItem is the shape returned by GET /promotions.
	PromotionListItem struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		StoreID   int    `json:"storeId"`
		StoreName string `json:"storeName"`
		StartsAt  string `json:"startsAt"`
		EndsAt    string `json:"endsAt"`
		IsActive  bool   `json:"isActive"`
		CreatedAt string `json:"createdAt"`
	}

	// PromotionPriceItem is one product's special price within a promotion.
	PromotionPriceItem struct {
		ProductID   int     `json:"productId"`
		ProductName string  `json:"productName"`
		Price       float64 `json:"price"`
	}

	// PromotionDetailResponse is the shape returned by POST /promotions and
	// GET /promotions/:id.
	PromotionDetailResponse struct {
		ID        int                   `json:"id"`
		Name      string                `json:"name"`
		StoreID   int                   `json:"storeId"`
		StoreName string                `json:"storeName"`
		StartsAt  string                `json:"startsAt"`
		EndsAt    string                `json:"endsAt"`
		IsActive  bool                  `json:"isActive"`
		Items     []PromotionPriceItem  `json:"items"`
		CreatedAt string                `json:"createdAt"`
		UpdatedAt string                `json:"updatedAt"`
	}
)

func (PromotionListItem) Collection(promotions []models.Promotion) []PromotionListItem {
	now := time.Now()
	items := make([]PromotionListItem, 0, len(promotions))
	for _, promotion := range promotions {
		storeName := ""
		if promotion.Store != nil {
			storeName = promotion.Store.Name
		}

		items = append(items, PromotionListItem{
			ID:        promotion.ID,
			Name:      promotion.Name,
			StoreID:   promotion.StoreID,
			StoreName: storeName,
			StartsAt:  promotion.StartsAt.Format("2006-01-02 15:04:05"),
			EndsAt:    promotion.EndsAt.Format("2006-01-02 15:04:05"),
			IsActive:  !now.Before(promotion.StartsAt) && !now.After(promotion.EndsAt),
			CreatedAt: promotion.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items
}

func (response *PromotionDetailResponse) Make(promotion models.Promotion) *PromotionDetailResponse {
	now := time.Now()

	storeName := ""
	if promotion.Store != nil {
		storeName = promotion.Store.Name
	}

	items := make([]PromotionPriceItem, 0, len(promotion.Prices))
	for _, price := range promotion.Prices {
		productName := ""
		if price.Product != nil {
			productName = price.Product.Name
		}
		items = append(items, PromotionPriceItem{
			ProductID:   price.ProductID,
			ProductName: productName,
			Price:       price.Price,
		})
	}

	return &PromotionDetailResponse{
		ID:        promotion.ID,
		Name:      promotion.Name,
		StoreID:   promotion.StoreID,
		StoreName: storeName,
		StartsAt:  promotion.StartsAt.Format("2006-01-02 15:04:05"),
		EndsAt:    promotion.EndsAt.Format("2006-01-02 15:04:05"),
		IsActive:  !now.Before(promotion.StartsAt) && !now.After(promotion.EndsAt),
		Items:     items,
		CreatedAt: promotion.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: promotion.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
