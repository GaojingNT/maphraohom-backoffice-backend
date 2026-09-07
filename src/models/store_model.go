package models

import "gorm.io/gorm"

type Store struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	// Logo is the MinIO object key only — never a full URL.
	Logo string `json:"logo" gorm:"column:logo;size:255;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Store searchable attributes
func StoreSearchable() []string {
	return []string{"name"}
}
