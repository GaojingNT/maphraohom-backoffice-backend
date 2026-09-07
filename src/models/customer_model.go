package models

import "gorm.io/gorm"

type Customer struct {
	BaseModel

	// Fields
	Name  string `json:"name" gorm:"column:name;size:255;not null;"`
	Phone string `json:"phone" gorm:"column:phone;size:50;"`

	// Address is the customer's latest/default address, used only to prefill
	// new bills. It is not linked to past bills — see Bill.CustomerAddress
	// for the snapshot taken at issue time.
	Address string `json:"address" gorm:"column:address;size:255;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Customer searchable attributes
func CustomerSearchable() []string {
	return []string{"name", "phone"}
}
