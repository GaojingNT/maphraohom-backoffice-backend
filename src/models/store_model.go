package models

import "gorm.io/gorm"

type Store struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	// Logo and Signature are MinIO object keys only — never a full URL.
	Logo      string `json:"logo" gorm:"column:logo;size:255;"`
	Signature string `json:"signature" gorm:"column:signature;size:255;"`

	Address string `json:"address" gorm:"column:address;size:255;"`
	Phone   string `json:"phone" gorm:"column:phone;size:50;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Store searchable attributes
func StoreSearchable() []string {
	return []string{"name"}
}
