package models

import "gorm.io/gorm"

type RoleGroup struct {
	BaseModel
	Name        string `json:"name" gorm:"column:role_group_name"`
	Code        string `json:"code" gorm:"column:role_group_code"`
	Description string `json:"description"`

	// Relations
	Roles []Role `json:"roles,omitempty" gorm:"many2many:role_groups_roles;"`

	// Audit fields
	CreatedByID *int  `json:"createdById,omitempty" gorm:"column:created_by"`
	CreatedBy   *User `json:"createdBy,omitempty"`
	UpdatedByID *int  `json:"updatedById,omitempty" gorm:"column:updated_by"`
	UpdatedBy   *User `json:"updatedBy,omitempty"`

	// Non database fields
	Users     []User `json:"users,omitempty" gorm:"-"`
	UserCount int    `json:"userCount" gorm:"-"`

	// Soft deletes
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

func (m *RoleGroup) AfterFind(tx *gorm.DB) (err error) {
	// Check if "Roles.Users" is preloaded
	if tx.Statement.Preloads != nil {
		if _, ok := tx.Statement.Preloads["Roles.Users"]; ok {
			for _, role := range m.Roles {
				m.Users = append(m.Users, role.Users...)
				m.UserCount += len(role.Users)
			}
		}
	}

	return
}

type RoleGroupRole struct {
	RoleGroupID int `json:"roleGroupID"`
	RoleID      int `json:"roleID"`
}

func (m *RoleGroupRole) TableName() string {
	return "tbl_role_groups_roles"
}
