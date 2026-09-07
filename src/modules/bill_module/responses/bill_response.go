package responses

import "maphraohom.app/maphraohom-backoffice/src/models"

type (
	// BillListItem is the shape returned by GET /bills (paginated list).
	BillListItem struct {
		ID              int     `json:"id"`
		CustomerName    string  `json:"customerName"`
		CustomerAddress string  `json:"customerAddress"`
		Total           float64 `json:"total"`
	}

	// BillDetailResponse is the shape returned by GET /bills/:id — every
	// bills table column except deleted_at.
	BillDetailResponse struct {
		ID              int     `json:"id"`
		ProductID       int     `json:"productId"`
		StoreID         int     `json:"storeId"`
		CustomerID      *int    `json:"customerId,omitempty"`
		BookNo          int     `json:"bookNo"`
		ReceiptNo       int     `json:"receiptNo"`
		CustomerName    string  `json:"customerName"`
		CustomerAddress string  `json:"customerAddress"`
		Kilogram        float64 `json:"kilogram"`
		Price           float64 `json:"price"`
		Discount        float64 `json:"discount"`
		ShippingFee     float64 `json:"shippingFee"`
		Total           float64 `json:"total"`
		Slip            string  `json:"slip"`
		CreatedAt       string  `json:"createdAt"`
		UpdatedAt       string  `json:"updatedAt"`
	}
)

func (BillListItem) Collection(bills []models.Bill) []BillListItem {
	items := make([]BillListItem, 0, len(bills))
	for _, bill := range bills {
		items = append(items, BillListItem{
			ID:              bill.ID,
			CustomerName:    bill.CustomerName,
			CustomerAddress: bill.CustomerAddress,
			Total:           bill.Total,
		})
	}
	return items
}

func (response *BillDetailResponse) Make(bill models.Bill) *BillDetailResponse {
	return &BillDetailResponse{
		ID:              bill.ID,
		ProductID:       bill.ProductID,
		StoreID:         bill.StoreID,
		CustomerID:      bill.CustomerID,
		BookNo:          bill.BookNo,
		ReceiptNo:       bill.ReceiptNo,
		CustomerName:    bill.CustomerName,
		CustomerAddress: bill.CustomerAddress,
		Kilogram:        bill.Kilogram,
		Price:           bill.Price,
		Discount:        bill.Discount,
		ShippingFee:     bill.ShippingFee,
		Total:           bill.Total,
		Slip:            bill.Slip,
		CreatedAt:       bill.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       bill.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
