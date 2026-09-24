package dtos

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreateBillItem is one line item of a create/update-bill request. Price is
// entered by the user — there is no price list to look up. Server computes
// Unit (from the product) and Subtotal; anything the client sends for those
// is ignored.
type CreateBillItem struct {
	ProductID int             `json:"productId"`
	Quantity  decimal.Decimal `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
}

// CreateBill is the JSON body of POST /api/v1/bills. Type, StoreID, and
// Items are required; CustomerID/Discount/ShippingFee default to
// nil/0/0 when omitted. Server computes Unit/Subtotal/Total/BookNo/ReceiptNo
// — anything the client sends for those fields is ignored.
type CreateBill struct {
	Type            string           `json:"type"`
	StoreID         int              `json:"storeId"`
	CustomerID      *int             `json:"customerId,omitempty"`
	CustomerName    string           `json:"customerName"`
	CustomerAddress string           `json:"customerAddress"`
	CustomerPhone   string           `json:"customerPhone"`
	Discount        decimal.Decimal  `json:"discount"`
	ShippingFee     decimal.Decimal  `json:"shippingFee"`
	Items           []CreateBillItem `json:"items"`
	// CreatedAt lets the client back- or post-date a bill (e.g. entering one
	// for a sale that actually happened earlier) — RFC3339, e.g.
	// "2026-09-20T14:30:00+07:00". Optional; defaults to the time the
	// request is processed when omitted. Only book/receipt numbering's
	// calendar-year rollover uses the real server clock — this field never
	// affects that.
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}

// UpdateBill is the JSON body of PUT /api/v1/bills/:id — same shape as
// CreateBill. StoreID and Type must match the existing bill's values (the
// service rejects the request with 400 otherwise); Items replace the bill's
// entire line-item set. BookNo/ReceiptNo and the slip are never touched.
type UpdateBill struct {
	Type            string           `json:"type"`
	StoreID         int              `json:"storeId"`
	CustomerID      *int             `json:"customerId,omitempty"`
	CustomerName    string           `json:"customerName"`
	CustomerAddress string           `json:"customerAddress"`
	CustomerPhone   string           `json:"customerPhone"`
	Discount        decimal.Decimal  `json:"discount"`
	ShippingFee     decimal.Decimal  `json:"shippingFee"`
	Items           []CreateBillItem `json:"items"`
}
