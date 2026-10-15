package models

import (
	"strings"

	"gorm.io/gorm"
)

const (
	// ProductScopeAll products (the coconut catalog) can go on any bill.
	ProductScopeAll = "all"
	// ProductScopePayment products can only go on a payment bill — they are
	// the items users typed in themselves on a payment bill.
	ProductScopePayment = "payment"
)

// CommonUnits are the units the payment-bill form offers first; the first
// is the default. Any other unit can be typed in (up to MaxUnitLength).
var CommonUnits = []string{"รายการ", "ครั้ง", "ชิ้น", "ถุง", "ลัง", "กล่อง", "แพ็ค", "กก.", "ขวด", "ลิตร", "เที่ยว", "วัน"}

// MaxUnitLength matches the size of tbl_products.unit / tbl_bill_items.unit.
const MaxUnitLength = 20

// NormalizeProductName trims a typed product name and collapses inner runs
// of whitespace, so "ค่า  ขนส่ง " and "ค่า ขนส่ง" are the same product.
// Units are normalized the same way.
func NormalizeProductName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

type Product struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	// Unit is the selling unit, e.g. "กก." or "ขวด".
	Unit string `json:"unit" gorm:"column:unit;size:20;not null;"`

	// BillScope is ProductScopeAll or ProductScopePayment — which bill types
	// may use this product.
	BillScope string `json:"billScope" gorm:"column:bill_scope;size:20;not null;default:all;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Product searchable attributes
func ProductSearchable() []string {
	return []string{"name"}
}
