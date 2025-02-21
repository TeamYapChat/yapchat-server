package repositories

import (
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	Update(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByVerificationCode(code string) (*models.User, error)
	Delete(id uint) error
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

func (r *MySQLUserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error

	return &user, err
}

func (r *MySQLUserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error

	return &user, err
}

func (r *MySQLUserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error

	return &user, err
}

func (r *MySQLUserRepository) FindByVerificationCode(code string) (*models.User, error) {
	var user models.User
	err := r.db.Where("verification_code = ?", code).First(&user).Error
	return &user, err
}

func (r *MySQLUserRepository) Delete(id uint) error {
	var user models.User
	err := r.db.Where("id = ?", id).Delete(&user).Error

	return err
}
