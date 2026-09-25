package models

// UserStore is the join row behind User.Stores / Store.Owners. Declared
// explicitly (rather than left to GORM's implicit many2many table) so the
// composite primary key and the reverse-lookup index on store_id are part
// of the schema — see SetupJoinTable in pkg/database.
type UserStore struct {
	UserID  int `json:"userId" gorm:"column:user_id;primaryKey;"`
	StoreID int `json:"storeId" gorm:"column:store_id;primaryKey;index;"`
}
