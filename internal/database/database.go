package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
)

var DB *gorm.DB

func Connect(dsn string) error {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	return nil
}

func Migrate() error {
	if err := DB.AutoMigrate(&models.User{}); err != nil {
		return err
	}

	return nil
}
