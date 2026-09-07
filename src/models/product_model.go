package models

import "gorm.io/gorm"

type Product struct {
	BaseModel

	// Fields
	Name string `json:"name" gorm:"column:name;size:255;not null;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"column:deleted_at;index;"`
}

// Product searchable attributes
func ProductSearchable() []string {
	return []string{"name"}
}
