package repositories

import (
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
)

type ChatRoomRepository struct {
	db *gorm.DB
}

func NewChatRoomRepository(db *gorm.DB) *ChatRoomRepository {
	return &ChatRoomRepository{db: db}
}

func (r *ChatRoomRepository) Create(chatroom *models.ChatRoom) error {
	return r.db.Create(chatroom).Error
}

func (r *ChatRoomRepository) GetByID(id uint) (*models.ChatRoom, error) {
	var chatroom models.ChatRoom
	err := r.db.Preload("Participants").First(&chatroom, id).Error
	if err != nil {
		return nil, err
	}
	return &chatroom, nil
}

func (r *ChatRoomRepository) List(userID uint) ([]*models.ChatRoom, error) {
	var chatrooms []*models.ChatRoom
	err := r.db.Preload("Participants").
		Joins("JOIN chat_room_participants ON chat_rooms.id = chat_room_participants.chat_room_id").
		Where("chat_room_participants.user_id = ?", userID).
		Find(&chatrooms).Error
	if err != nil {
		return nil, err
	}
	return chatrooms, nil
}

func (r *ChatRoomRepository) Update(chatroom *models.ChatRoom) error {
	return r.db.Save(chatroom).Error
}

func (r *ChatRoomRepository) Delete(id uint) error {
	return r.db.Delete(&models.ChatRoom{}, id).Error
}

func (r *ChatRoomRepository) AddParticipant(chatroomID uint, userID uint) error {
	chatroom, err := r.GetByID(chatroomID)
	if err != nil {
		return err
	}
	user := &models.User{Model: gorm.Model{ID: userID}}
	return r.db.Model(chatroom).Association("Participants").Append(user)
}

func (r *ChatRoomRepository) RemoveParticipant(chatroomID uint, userID uint) error {
	chatroom, err := r.GetByID(chatroomID)
	if err != nil {
		return err
	}
	user := &models.User{Model: gorm.Model{ID: userID}}
	return r.db.Model(chatroom).Association("Participants").Delete(user)
}
