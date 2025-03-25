package repositories

import (
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	Update(user *models.User) error
	UpdateStatus(user *models.User) error
	UpdateImage(user *models.User) error
	UpdateUsername(user *models.User) error
	FindByID(id string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	Delete(id string) error
}

type MySQLUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *MySQLUserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *MySQLUserRepository) UpdateImage(user *models.User) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("image_url", user.ImageURL).
		Error
}

func (r *MySQLUserRepository) UpdateStatus(user *models.User) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("is_online", user.IsOnline).
		Error
}

func (r *MySQLUserRepository) UpdateUsername(user *models.User) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Update("username", user.Username).
		Error
}

func (r *MySQLUserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error

	return &user, err
}

func (r *MySQLUserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error

	return &user, err
}

func (r *MySQLUserRepository) Delete(id string) error {
	var user models.User
	err := r.db.Where("id = ?", id).Delete(&user).Error

	return err
}
