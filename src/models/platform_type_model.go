package models

// PlatformType is a sales channel a receipt bill can be tagged with
// (Facebook, TikTok, Lazada, Shopee, Thaimart, ...). Rows are seeded on app
// start (pkg/database/platform_type_seed.go) and referenced by
// Bill.PlatformTypeID — a platform that any bill uses can't be deleted
// (FK ON DELETE RESTRICT); set IsActive false to hide it from the picker
// instead.
type PlatformType struct {
	BaseModel

	// Code is the stable machine name, e.g. "shopee" — unique.
	Code string `json:"code" gorm:"column:code;size:30;not null;uniqueIndex;"`
	// Name is what the UI shows, e.g. "Shopee".
	Name string `json:"name" gorm:"column:name;size:100;not null;"`
	// Logo is a storage object key only — never a full URL (like Store.Logo).
	Logo string `json:"logo" gorm:"column:logo;size:255;"`
	// SortOrder is the order the picker and filters list platforms in.
	SortOrder int `json:"sortOrder" gorm:"column:sort_order;not null;default:0;"`
	// IsActive false hides the platform from GET /platform-types (new
	// bills); bills already tagged with it keep showing it.
	IsActive bool `json:"isActive" gorm:"column:is_active;not null;default:true;"`
}
