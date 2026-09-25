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

	// Signature is a MinIO object key only — never a full URL. It's printed
	// on every receipt this user exports (moved here from Store, so the
	// signature belongs to the person issuing the document, not the shop).
	Signature string `json:"signature" gorm:"column:signature;size:255;"`

	// User has only one role
	RoleID *int  `json:"roleId,omitempty" gorm:"column:role_id;null;"`
	Role   *Role `json:"role,omitempty" gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	// Stores this user owns — many-to-many through tbl_user_stores (a store
	// can have several owners, and an owner several stores).
	Stores []Store `json:"stores,omitempty" gorm:"many2many:user_stores;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

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
