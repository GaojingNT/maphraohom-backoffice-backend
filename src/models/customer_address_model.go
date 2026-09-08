package models

import "gorm.io/gorm"

type CustomerAddress struct {
	BaseModel

	// Relations
	CustomerID int       `json:"customerId" gorm:"column:customer_id;not null;uniqueIndex:idx_customer_addresses_one_default,where:is_default = true;"`
	Customer   *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Fields
	Address string `json:"address" gorm:"column:address;size:255;not null;"`

	// Label is an optional free-form tag, e.g. "บ้าน", "สวน", "จุดรับของ 2".
	Label string `json:"label,omitempty" gorm:"column:label;size:100;"`

	// IsDefault marks the address used to prefill new bills. Only one
	// address per customer may be default — enforced by the partial unique
	// index above (customer_id, where is_default = true) and mirrored by
	// the application whenever a new address is added.
	IsDefault bool `json:"isDefault" gorm:"column:is_default;not null;default:false;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}
