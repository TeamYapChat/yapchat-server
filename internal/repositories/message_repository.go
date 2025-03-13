package repositories

import (
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(message *models.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) GetByRoomID(roomID uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("room_id = ? AND type = ?", roomID, "message").
		Order("created_at desc").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}
