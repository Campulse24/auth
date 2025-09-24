package models

import (
	"time"

	"gorm.io/gorm"
)

type VerificationToken struct {
	gorm.Model
	UserID uint
	Token  string    `gorm:"unique;not null"`
	Expiry time.Time `gorm:"not null"`
}
