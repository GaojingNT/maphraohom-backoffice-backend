package models

import "gorm.io/gorm"

type CustomerPhone struct {
	BaseModel

	CustomerID int       `json:"customerId" gorm:"column:customer_id;not null;uniqueIndex:idx_customer_phones_one_default,where:is_default = true AND deleted_at IS NULL;"`
	Customer   *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Phone string `json:"phone" gorm:"column:phone;size:50;not null;"`

	// Label is an optional free-form tag, e.g. "มือถือ", "ไลน์".
	Label string `json:"label,omitempty" gorm:"column:label;size:100;"`

	// IsDefault marks the phone used to prefill new bills. Only one phone
	// per customer may be default — enforced by the partial unique index above.
	IsDefault bool `json:"isDefault" gorm:"column:is_default;not null;default:false;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}
