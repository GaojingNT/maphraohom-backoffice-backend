package models

// MaxSlipsPerBill caps how many slip images one bill can carry.
const MaxSlipsPerBill = 5

// BillSlip is one slip image attached to a bill. A bill has 0..MaxSlipsPerBill
// of them, shown in id order (the order they were attached).
type BillSlip struct {
	BaseModel

	BillID int `json:"billId" gorm:"column:bill_id;not null;index;"`

	// ObjectKey is the MinIO object key only — never a full URL.
	ObjectKey string `json:"objectKey" gorm:"column:object_key;size:255;not null;"`

	// UploadedBy is the user logged in when the slip was attached. Nil for
	// slips carried over from bills.slip (who uploaded them was never
	// recorded).
	UploadedBy *int  `json:"uploadedBy" gorm:"column:uploaded_by;null;"`
	Uploader   *User `json:"-" gorm:"foreignKey:UploadedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
