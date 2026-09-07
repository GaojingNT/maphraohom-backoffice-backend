package models

import "gorm.io/gorm"

type Bill struct {
	BaseModel

	// Relations
	ProductID int      `json:"productId" gorm:"column:product_id;not null;"`
	Product   *Product `json:"product,omitempty" gorm:"foreignKey:ProductID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	StoreID int    `json:"storeId" gorm:"column:store_id;not null;"`
	Store   *Store `json:"store,omitempty" gorm:"foreignKey:StoreID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// Optional — used for cumulative purchase reporting only.
	CustomerID *int      `json:"customerId,omitempty" gorm:"column:customer_id;null;"`
	Customer   *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	BookNo    int `json:"bookNo" gorm:"column:book_no;not null;"`
	ReceiptNo int `json:"receiptNo" gorm:"column:receipt_no;not null;"`

	// Snapshots taken at the time the bill was issued — intentionally not
	// kept in sync with Customer, since old bills must keep the address
	// that was valid when they were issued.
	CustomerName    string `json:"customerName" gorm:"column:customer_name;size:255;"`
	CustomerAddress string `json:"customerAddress" gorm:"column:customer_address;size:255;"`

	Kilogram float64 `json:"kilogram" gorm:"column:kilogram;not null;"`

	// Price snapshot copied from StoreProductPrice at issue time.
	Price float64 `json:"price" gorm:"column:price;not null;"`

	Discount    float64 `json:"discount" gorm:"column:discount;not null;default:0;"`
	ShippingFee float64 `json:"shippingFee" gorm:"column:shipping_fee;not null;default:0;"`

	// Total = kilogram*price - discount + shippingFee
	Total float64 `json:"total" gorm:"column:total;not null;"`

	// Slip is the MinIO object key only — never a full URL.
	Slip string `json:"slip" gorm:"column:slip;size:255;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Bill searchable attributes
func BillSearchable() []string {
	return []string{"customer_name", "customer_address"}
}
