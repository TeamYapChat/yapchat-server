package database

import (
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

func CreateUser(email, username, password string) error {
	user := models.User{
		Email:             email,
		Username:          username,
		Password:          password,
		VerificationToken: utils.GenerateToken(),
	}

	if err := user.HashPassword(); err != nil {
		return err
	}

	result := DB.Create(&user)
	if result.Error != nil {
		return result.Error
	}

	if err := utils.SendVerificationEmail(user); err != nil {
		return err
	}

	return nil
}

func VerifyUser(id uint) error {
	var user models.User
	result := DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return result.Error
	}

	user.IsVerified = true
	user.VerificationToken = utils.HashUserID(id)
	DB.Save(&user)

	return nil
}

func GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func GetUserByVerificationToken(token string) (*models.User, error) {
	var user models.User
	result := DB.Where("verification_token = ?", token).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
