package models

import (
	"database/sql"

	"gorm.io/gorm"
)

type User struct {
	BaseModel
	// BaseAuditorModel // For user management, we need to know who created and updated the user

	// Fields
	Email                  string        `json:"email" gorm:"column:email;size:100;uniqueIndex;not null;"`
	FirstName              string        `json:"firstName" gorm:"column:first_name;size:100;"`
	LastName               string        `json:"lastName" gorm:"column:last_name;size:100;"`
	Password               string        `json:"-" gorm:"column:password;type:text;"`
	ResetPasswordToken     string        `json:"-" gorm:"column:reset_password_token;type:text;"`
	ResetPasswordExpiredAt *sql.NullTime `json:"-" gorm:"column:reset_password_expired_at;type:timestamp;"`

	// User has only one role
	RoleID *int  `json:"roleId,omitempty" gorm:"column:role_id;null;"`
	Role   *Role `json:"role,omitempty" gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Soft delete
	DeletedAt gorm.DeletedAt `json:"-" gorm:"column:deleted_at;index;"`
}

// Check if user has role
func (m *User) HasRole(role string) bool {
	// Check preload
	if m.Role == nil {
		return false
	}

	return m.Role.Name == role
}

// Check if user has permission
func (m *User) HasPermission(permission string) bool {
	// Check preload
	if m.Role == nil {
		return false
	}

	for _, rolePermission := range m.Role.Permissions {
		if rolePermission.Name == permission {
			return true
		}
	}

	return false
}

// User searchable attributes
func UserSearchable() []string {
	return []string{"email", "first_name", "last_name"}
}
