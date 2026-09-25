package models

import "gorm.io/gorm"

type Store struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	// Logo is a MinIO object key only — never a full URL. (The signature
	// used to live here too; it's now per user — see User.Signature.)
	Logo string `json:"logo" gorm:"column:logo;size:255;"`

	Address string `json:"address" gorm:"column:address;size:255;"`
	Phone   string `json:"phone" gorm:"column:phone;size:50;"`

	// Owners — see User.Stores.
	Owners []User `json:"owners,omitempty" gorm:"many2many:user_stores;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Store searchable attributes
func StoreSearchable() []string {
	return []string{"name"}
}
