package models

import "gorm.io/gorm"

type Customer struct {
	BaseModel

	// Fields
	Name  string `json:"name" gorm:"column:name;size:255;not null;"`
	Phone string `json:"phone" gorm:"column:phone;size:50;"`

	// A customer can have many addresses; CustomerAddress.IsDefault marks
	// the one used to prefill new bills. Bill.CustomerAddress holds an
	// independent snapshot taken at issue time — it is not linked here.
	Addresses []CustomerAddress `json:"addresses,omitempty" gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// A customer can have many phone numbers; CustomerPhone.IsDefault marks
	// the one used to prefill new bills.
	Phones []CustomerPhone `json:"phones,omitempty" gorm:"foreignKey:CustomerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Customer searchable attributes
func CustomerSearchable() []string {
	return []string{"name", "phone"}
}
