package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	gorm.Model
	UserID    uint       `gorm:"not null;index"`
	User      User       `gorm:"foreignKey:UserID"`
	TokenHash string     `gorm:"not null"`
	Expiry    time.Time  `gorm:"not null"`
	RevokedAt *time.Time `gorm:"nullable"`
}
