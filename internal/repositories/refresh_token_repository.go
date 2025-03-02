package repositories

import (
	"time"

	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
)

type RefreshTokenRepository interface {
	Create(refreshToken *models.RefreshToken) error
	FindByTokenHash(tokenHash string) (*models.RefreshToken, error)
	FindByUserID(userID uint) (*models.RefreshToken, error)
	Revoke(refreshToken *models.RefreshToken) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db}
}

func (r *refreshTokenRepository) Create(refreshToken *models.RefreshToken) error {
	return r.db.Create(refreshToken).Error
}

func (r *refreshTokenRepository) FindByTokenHash(tokenHash string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("token_hash = ?", tokenHash).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) FindByUserID(userID uint) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("user_id = ? AND revoked_at IS NULL", userID).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) Revoke(refreshToken *models.RefreshToken) error {
	return r.db.Model(refreshToken).Update("RevokedAt", time.Now()).Error
}
