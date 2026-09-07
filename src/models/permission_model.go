package models

type GroupPermission struct {
	BaseModel
	Name string `json:"name" gorm:"column:name;size:100;uniqueIndex;not null;"`

	// Relations
	Permissions []Permission `json:"permissions,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Permission struct {
	BaseModel
	Name string `json:"name" gorm:"column:name;size:100;uniqueIndex;not null;"`

	// Relations
	GroupPermissionID *uint            `json:"groupPermissionId,omitempty" gorm:"column:group_permission_id;null;"`
	GroupPermission   *GroupPermission `json:"groupPermission,omitempty" gorm:"foreignKey:GroupPermissionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Roles []Role `json:"roles,omitempty" gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE;"`

	MenuID uint  `json:"menuId"`
	Menu   *Menu `json:"-"`
}
