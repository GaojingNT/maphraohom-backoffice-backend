package models

import "gorm.io/gorm"

type CustomerAddress struct {
	BaseModel

	// Relations
	CustomerID int       `json:"customerId" gorm:"column:customer_id;not null;"`
	Customer   *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Fields
	Address string `json:"address" gorm:"column:address;size:255;not null;"`

	// IsDefault marks the address used to prefill new bills. It is not
	// linked to past bills — see Bill.CustomerAddress for the snapshot
	// taken at issue time.
	IsDefault bool `json:"isDefault" gorm:"column:is_default;not null;default:false;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}
