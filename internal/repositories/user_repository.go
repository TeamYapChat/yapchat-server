package repositories

import (
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	FindUserByEmail(email string) (*models.User, error)
	FindUserByVerificationCode(code string) (*models.User, error)
}

type MySQLUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *MySQLUserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *MySQLUserRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error

	return &user, err
}

func (r *MySQLUserRepository) FindUserByVerificationCode(code string) (*models.User, error) {
	var user models.User
	err := r.db.Where("verification_code = ?", code).First(&user).Error

	return &user, err
}
