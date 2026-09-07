package models

import "gorm.io/gorm"

type Role struct {
	BaseModel
	// BaseAuditorModel // For role management, we need to know who created and updated the role

	// Fields
	Name        string `json:"name" gorm:"column:name;size:100;uniqueIndex;not null"`
	Label       string `json:"label" gorm:"column:label;size:100;null;"`
	Description string `json:"description" gorm:"column:description;type:text;null;"`

	// Relations
	RoleGroups  []RoleGroup  `json:"roleGroups" gorm:"many2many:role_groups_roles;"`
	Permissions []Permission `json:"permissions,omitempty" gorm:"many2many:roles_permissions;constraint:OnDelete:CASCADE;"`
	Users       []User       `json:"users,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Menus       []Menu       `json:"menus" gorm:"many2many:roles_menus;"`

	// Audit fields
	CreatedByID *int  `json:"createdById,omitempty" gorm:"column:created_by"`
	CreatedBy   *User `json:"createdBy,omitempty"`
	UpdatedByID *int  `json:"updatedById,omitempty" gorm:"column:updated_by"`
	UpdatedBy   *User `json:"updatedBy,omitempty"`

	// Soft deletes
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

type RolePermission struct {
	RoleID       int `json:"roleID"`
	PermissionID int `json:"permissionID"`
}

func (m *RolePermission) TableName() string {
	return "roles_permissions"
}

type RoleMenu struct {
	RoleID int `json:"roleID"`
	MenuID int `json:"menuID"`
}

func (m *RoleMenu) TableName() string {
	return "roles_menus"
}
