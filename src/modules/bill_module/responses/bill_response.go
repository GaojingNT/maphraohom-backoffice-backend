package responses

import "maphraohom.app/maphraohom-backoffice/src/models"

type (
	// BillListItem is the shape returned by GET /bills (paginated list).
	BillListItem struct {
		ID              int     `json:"id"`
		StoreID         int     `json:"storeId"`
		StoreName       string  `json:"storeName"`
		ReceiptNo       int     `json:"receiptNo"`
		CustomerName    string  `json:"customerName"`
		CustomerAddress string  `json:"customerAddress"`
		Total           float64 `json:"total"`
		TotalKilogram   float64 `json:"totalKilogram"`
		ItemCount       int     `json:"itemCount"`
		CreatedAt       string  `json:"createdAt"`
	}

	// BillItemDetail is one line item within BillDetailResponse.
	BillItemDetail struct {
		ID          int     `json:"id"`
		ProductID   int     `json:"productId"`
		ProductName string  `json:"productName"`
		Kilogram    float64 `json:"kilogram"`
		Price       float64 `json:"price"`
		Subtotal    float64 `json:"subtotal"`
	}

	// BillDetailResponse is the shape returned by GET /bills/:id — every
	// bills table column except deleted_at, plus store name/logo and its
	// line items (with each item's product name).
	BillDetailResponse struct {
		ID              int              `json:"id"`
		StoreID         int              `json:"storeId"`
		StoreName       string           `json:"storeName"`
		StoreLogo       string           `json:"storeLogo"`
		CustomerID      *int             `json:"customerId,omitempty"`
		BookNo          int              `json:"bookNo"`
		ReceiptNo       int              `json:"receiptNo"`
		CustomerName    string           `json:"customerName"`
		CustomerAddress string           `json:"customerAddress"`
		Discount        float64          `json:"discount"`
		ShippingFee     float64          `json:"shippingFee"`
		Total           float64          `json:"total"`
		Slip            string           `json:"slip"`
		CreatedAt       string           `json:"createdAt"`
		UpdatedAt       string           `json:"updatedAt"`
		Items           []BillItemDetail `json:"items"`
	}
)

func (BillListItem) Collection(bills []models.Bill) []BillListItem {
	items := make([]BillListItem, 0, len(bills))
	for _, bill := range bills {
		var totalKilogram float64
		for _, item := range bill.Items {
			totalKilogram += item.Kilogram
		}

		storeName := ""
		if bill.Store != nil {
			storeName = bill.Store.Name
		}

		items = append(items, BillListItem{
			ID:              bill.ID,
			StoreID:         bill.StoreID,
			StoreName:       storeName,
			ReceiptNo:       bill.ReceiptNo,
			CustomerName:    bill.CustomerName,
			CustomerAddress: bill.CustomerAddress,
			Total:           bill.Total,
			TotalKilogram:   totalKilogram,
			ItemCount:       len(bill.Items),
			CreatedAt:       bill.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items
}

func (response *BillDetailResponse) Make(bill models.Bill) *BillDetailResponse {
	items := make([]BillItemDetail, 0, len(bill.Items))
	for _, item := range bill.Items {
		productName := ""
		if item.Product != nil {
			productName = item.Product.Name
		}

		items = append(items, BillItemDetail{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: productName,
			Kilogram:    item.Kilogram,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		})
	}

	var storeName, storeLogo string
	if bill.Store != nil {
		storeName = bill.Store.Name
		storeLogo = bill.Store.Logo
	}

	return &BillDetailResponse{
		ID:              bill.ID,
		StoreID:         bill.StoreID,
		StoreName:       storeName,
		StoreLogo:       storeLogo,
		CustomerID:      bill.CustomerID,
		BookNo:          bill.BookNo,
		ReceiptNo:       bill.ReceiptNo,
		CustomerName:    bill.CustomerName,
		CustomerAddress: bill.CustomerAddress,
		Discount:        bill.Discount,
		ShippingFee:     bill.ShippingFee,
		Total:           bill.Total,
		Slip:            bill.Slip,
		CreatedAt:       bill.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       bill.UpdatedAt.Format("2006-01-02 15:04:05"),
		Items:           items,
	}
}
