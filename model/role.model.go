package model

type Role struct {
	ID Decimal `gorm:"uniqueIndex" json:"id"`
}
