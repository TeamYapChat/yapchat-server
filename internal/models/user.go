package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           string  `gorm:"primarykey;varchar(255)"`
	Username     string  `gorm:"uniqueIndex;not null;type:varchar(24)"`
	ImageURL     string  `gorm:"varchar(100)"`
	IsOnline     bool    `gorm:"default:false"`
	BlockedUsers []*User `gorm:"many2many:blocked_users"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
