package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username         string `gorm:"uniqueIndex;not null;type:varchar(24)"`
	Email            string `gorm:"uniqueIndex;not null;type:varchar(100)"`
	Password         string `gorm:"not null;type:varchar(64)"`
	IsVerified       bool   `gorm:"default:false"`
	VerificationCode string `gorm:"size:255"`
	ImageURL         string `gorm:"varchar(100)"`
	IsOnline         bool   `gorm:"default:false"`
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		u.Password = string(hashedPassword)
	}

	return nil
}
