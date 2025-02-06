package database

import "github.com/teamyapchat/yapchat-server/internal/models"

func CreateUser(username, password string) error {
	user := models.User{Username: username, Password: password}

	if err := user.HashPassword(); err != nil {
		return err
	}

	result := DB.Create(&user)
	return result.Error
}

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
