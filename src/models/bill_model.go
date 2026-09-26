package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	// BillTypeReceipt is a receipt bill ("บิลรับเงิน") — money coming in
	// from a customer.
	BillTypeReceipt = "receipt"
	// BillTypePayment is a payment bill ("บิลจ่ายเงิน") — money going out
	// to a payee.
	BillTypePayment = "payment"
)

// IsValidBillType reports whether t is one of the whitelisted bill types.
// Validate every bills.type value (request input and query filters) against
// this rather than any other check.
func IsValidBillType(t string) bool {
	return t == BillTypeReceipt || t == BillTypePayment
}

type Bill struct {
	BaseModel

	// Relations
	StoreID int    `json:"storeId" gorm:"column:store_id;not null;"`
	Store   *Store `json:"store,omitempty" gorm:"foreignKey:StoreID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Type is 'receipt' or 'payment' — set once at creation and never
	// changed afterwards (book/receipt numbers are already tied to this
	// type's sequence). Every other field is shared by both kinds of bill.
	Type string `json:"type" gorm:"column:type;size:20;not null;default:receipt;"`

	// Optional — used for cumulative purchase reporting only.
	CustomerID *int      `json:"customerId,omitempty" gorm:"column:customer_id;null;"`
	Customer   *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Line items — one bill can carry several products.
	Items []BillItem `json:"items,omitempty" gorm:"foreignKey:BillID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	BookNo    int `json:"bookNo" gorm:"column:book_no;not null;"`
	ReceiptNo int `json:"receiptNo" gorm:"column:receipt_no;not null;"`

	// Snapshots taken at the time the bill was issued — intentionally not
	// kept in sync with Customer, since old bills must keep the values
	// that were valid when they were issued. "Customer" here really means
	// "counterparty": the customer on a receipt bill, or the payee on a
	// payment bill.
	CustomerName    string `json:"customerName" gorm:"column:customer_name;size:255;"`
	CustomerAddress string `json:"customerAddress" gorm:"column:customer_address;size:255;"`
	CustomerPhone   string `json:"customerPhone" gorm:"column:customer_phone;size:50;"`

	Discount    decimal.Decimal `json:"discount" gorm:"column:discount;type:numeric(12,2);not null;default:0;"`
	ShippingFee decimal.Decimal `json:"shippingFee" gorm:"column:shipping_fee;type:numeric(12,2);not null;default:0;"`

	// Total = SUM(bill_items.subtotal) - discount + shippingFee — always
	// positive; the direction of money is read from Type.
	Total decimal.Decimal `json:"total" gorm:"column:total;type:numeric(12,2);not null;"`

	// Slip is the MinIO object key only — never a full URL. Nil means no
	// slip has been attached yet; it is set/cleared only through the
	// dedicated slip endpoints, never by POST/PUT /bills.
	Slip *string `json:"slip" gorm:"column:slip;size:255;"`

	// Audit fields. updated_at can't answer "when was this bill edited" —
	// it also moves when a slip is attached or removed — so edits and slip
	// changes are tracked separately. All nil for bills from before these
	// columns existed (edited_at/created_by were never recorded).
	//
	// CreatedBy is the user logged in when the bill was created. Creator is
	// preloaded with id/first_name/last_name only and never serialized
	// straight from the model (it would carry email and signature).
	CreatedBy *int  `json:"createdBy" gorm:"column:created_by;null;"`
	Creator   *User `json:"-" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	// EditedAt is set by PUT /bills/:id only.
	EditedAt *time.Time `json:"editedAt" gorm:"column:edited_at;null;"`
	// SlipUploadedAt is set when a slip is attached or replaced and cleared
	// when it's removed — always together with Slip.
	SlipUploadedAt *time.Time `json:"slipUploadedAt" gorm:"column:slip_uploaded_at;null;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Bill searchable attributes
func BillSearchable() []string {
	return []string{"customer_name", "customer_address"}
}
