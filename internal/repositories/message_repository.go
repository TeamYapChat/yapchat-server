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

func (r *MessageRepository) GetByRoomID(roomID uint, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("room_id = ?", roomID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) GetCountByRoomID(roomID uint) (int, error) {
	var count int64
	err := r.db.Model(&models.Message{}).Where("room_id = ?", roomID).Count(&count).Error
	return int(count), err
}
