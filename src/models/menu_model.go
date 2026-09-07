package models

import "gorm.io/gorm"

type Menu struct {
	BaseModel
	Name        string `json:"name"`
	Label       string `json:"label"`
	KeyName     string `json:"keyName"`
	Icon        string `json:"icon"`
	UrlPath     string `json:"urlPath"`
	Description string `json:"description"`
	IsFeature   bool   `json:"isFeature"`
	IsActive    bool   `json:"isActive"`
	Mode        string `json:"mode"`

	// Relationships
	ParentID *int  `json:"parentId"`
	Parent   *Menu `json:"parent" gorm:"foreignKey:ParentID"`

	SubMenus []Menu `json:"subMenus" gorm:"foreignKey:ParentID"`

	Permissions []Permission `json:"permissions"`

	// Audit fields
	CreatedByID *int  `json:"createdById,omitempty" gorm:"column:created_by"`
	CreatedBy   *User `json:"createdBy,omitempty"`
	UpdatedByID *int  `json:"updatedById,omitempty" gorm:"column:updated_by"`
	UpdatedBy   *User `json:"updatedBy,omitempty"`

	// Soft deletes
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}
