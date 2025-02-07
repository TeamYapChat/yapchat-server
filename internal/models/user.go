package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username          string `gorm:"unique;not null;type:varchar(24)"`
	Password          string `gorm:"not null;type:varchar(64)"`
	Email             string `gorm:"uniqueIndex;not null;type:varchar(100)"`
	IsVerified        bool   `gorm:"default:false"`
	VerificationToken string `gorm:"uniqueIndex;not null;type:varchar(64)"`
}

func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
