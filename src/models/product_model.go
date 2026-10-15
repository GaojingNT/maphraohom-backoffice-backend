package models

import (
	"slices"
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

// PaymentItemUnits are the units a user can pick for an item they type in
// on a payment bill; the first is the default.
var PaymentItemUnits = []string{"รายการ", "ครั้ง", "ชิ้น", "ถุง", "กก.", "ขวด"}

func IsValidPaymentItemUnit(unit string) bool {
	return slices.Contains(PaymentItemUnits, unit)
}

// NormalizeProductName trims a typed product name and collapses inner runs
// of whitespace, so "ค่า  ขนส่ง " and "ค่า ขนส่ง" are the same product.
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
