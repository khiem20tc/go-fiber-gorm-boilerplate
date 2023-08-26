package model

type User struct {
	ID    string `gorm:"uniqueIndex" json:"id"`
	Email string `gorm:"uniqueIndex" json:"email"`
}