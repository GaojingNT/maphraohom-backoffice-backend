package responses

import (
	"fmt"

	"github.com/shopspring/decimal"
	"maphraohom.app/maphraohom-backoffice/src/models"
)

type (
	// BillListItem is the shape returned by GET /bills (paginated list).
	BillListItem struct {
		ID              int             `json:"id"`
		Type            string          `json:"type"`
		StoreID         int             `json:"storeId"`
		StoreName       string          `json:"storeName"`
		BookNo          int             `json:"bookNo"`
		ReceiptNo       int             `json:"receiptNo"`
		CustomerName    string          `json:"customerName"`
		CustomerAddress string          `json:"customerAddress"`
		Total           decimal.Decimal `json:"total"`
		ItemCount       int             `json:"itemCount"`
		HasSlip         bool            `json:"hasSlip"`
		CreatedAt       string          `json:"createdAt"`
	}

	// BillItemDetail is one line item within BillDetailResponse.
	BillItemDetail struct {
		ID          int             `json:"id"`
		ProductID   int             `json:"productId"`
		ProductName string          `json:"productName"`
		Quantity    decimal.Decimal `json:"quantity"`
		Unit        string          `json:"unit"`
		Price       decimal.Decimal `json:"price"`
		Subtotal    decimal.Decimal `json:"subtotal"`
	}

	// BillDetailResponse is the shape returned by GET /bills/:id — every
	// bills table column except deleted_at, plus store name/logo, a
	// ready-to-use slip URL, and its line items (with each item's product
	// name).
	BillDetailResponse struct {
		ID              int              `json:"id"`
		Type            string           `json:"type"`
		StoreID         int              `json:"storeId"`
		StoreName       string           `json:"storeName"`
		StoreLogo       string           `json:"storeLogo"`
		CustomerID      *int             `json:"customerId,omitempty"`
		BookNo          int              `json:"bookNo"`
		ReceiptNo       int              `json:"receiptNo"`
		CustomerName    string           `json:"customerName"`
		CustomerAddress string           `json:"customerAddress"`
		CustomerPhone   string           `json:"customerPhone"`
		Discount        decimal.Decimal  `json:"discount"`
		ShippingFee     decimal.Decimal  `json:"shippingFee"`
		Total           decimal.Decimal  `json:"total"`
		SlipURL         *string          `json:"slipUrl"`
		CreatedAt       string           `json:"createdAt"`
		UpdatedAt       string           `json:"updatedAt"`
		Items           []BillItemDetail `json:"items"`
	}
)

// SlipURLBuilder turns a bill's stored slip object key into a URL the
// client can load directly. Set once at startup (see bill_module.go) to the
// existing GET /api/v1/files/* proxy, which streams from whichever storage
// backend (local disk or S3/MinIO) is configured — bill_response never talks
// to storage directly.
var SlipURLBuilder = func(key string) string {
	return fmt.Sprintf("/api/v1/files/%s", key)
}

func (BillListItem) Collection(bills []models.Bill) []BillListItem {
	items := make([]BillListItem, 0, len(bills))
	for _, bill := range bills {
		storeName := ""
		if bill.Store != nil {
			storeName = bill.Store.Name
		}

		items = append(items, BillListItem{
			ID:              bill.ID,
			Type:            bill.Type,
			StoreID:         bill.StoreID,
			StoreName:       storeName,
			BookNo:          bill.BookNo,
			ReceiptNo:       bill.ReceiptNo,
			CustomerName:    bill.CustomerName,
			CustomerAddress: bill.CustomerAddress,
			Total:           bill.Total,
			ItemCount:       len(bill.Items),
			HasSlip:         bill.Slip != nil,
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
			Quantity:    item.Quantity,
			Unit:        item.Unit,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		})
	}

	var storeName, storeLogo string
	if bill.Store != nil {
		storeName = bill.Store.Name
		storeLogo = bill.Store.Logo
	}

	var slipURL *string
	if bill.Slip != nil {
		url := SlipURLBuilder(*bill.Slip)
		slipURL = &url
	}

	return &BillDetailResponse{
		ID:              bill.ID,
		Type:            bill.Type,
		StoreID:         bill.StoreID,
		StoreName:       storeName,
		StoreLogo:       storeLogo,
		CustomerID:      bill.CustomerID,
		BookNo:          bill.BookNo,
		ReceiptNo:       bill.ReceiptNo,
		CustomerName:    bill.CustomerName,
		CustomerAddress: bill.CustomerAddress,
		CustomerPhone:   bill.CustomerPhone,
		Discount:        bill.Discount,
		ShippingFee:     bill.ShippingFee,
		Total:           bill.Total,
		SlipURL:         slipURL,
		CreatedAt:       bill.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       bill.UpdatedAt.Format("2006-01-02 15:04:05"),
		Items:           items,
	}
}
