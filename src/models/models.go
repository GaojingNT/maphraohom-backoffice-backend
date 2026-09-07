package models

import (
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID int `json:"id" gorm:"column:id;primarykey;"`

	BaseTimestampModel
}

type BaseTimestampModel struct {
	CreatedAt time.Time `json:"createdAt,omitempty" gorm:"column:created_at;default:CURRENT_TIMESTAMP;not null;"`
	UpdatedAt time.Time `json:"updatedAt,omitempty" gorm:"column:updated_at;default:CURRENT_TIMESTAMP;not null;"`
}

type BaseAuditorModel struct {
	CreatedBy *int  `json:"createdBy,omitempty" gorm:"column:created_by;not null;"`
	Creator   *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	UpdatedBy *int  `json:"updatedBy,omitempty" gorm:"column:updated_by;not null;"`
	Updater   *User `json:"updater,omitempty" gorm:"foreignKey:UpdatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Searching scope
func SearchingScope(searchable []string, searchAttribute string, searchByAttribute string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if searchAttribute == "" {
			return db
		}

		if searchByAttribute != "" {
			if slices.Contains(searchable, searchByAttribute) {
				return db.Where(fmt.Sprintf("%s LIKE ?", searchByAttribute), "%"+searchAttribute+"%")
			}
			return db
		}

		if len(searchable) > 0 {
			query := ""
			values := make([]interface{}, len(searchable))
			for i, attribute := range searchable {
				if i != 0 {
					query += " OR "
				}
				query += fmt.Sprintf("%s LIKE ?", attribute)
				values[i] = "%" + searchAttribute + "%"
			}
			return db.Where(query, values...)
		}

		return db
	}
}
