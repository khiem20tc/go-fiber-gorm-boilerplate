package model

type Role struct {
	ID string `gorm:"uniqueIndex" json:"id"`
}
