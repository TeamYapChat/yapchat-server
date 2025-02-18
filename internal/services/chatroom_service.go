package services

import (
	"errors"

	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

var ErrChatRoomNotFound = errors.New("chat room not found")

type ChatRoomService struct {
	repo *repositories.ChatRoomRepository
}

func NewChatRoomService(repo *repositories.ChatRoomRepository) *ChatRoomService {
	return &ChatRoomService{repo: repo}
}

func (s *ChatRoomService) CreateChatRoom(chatroom *models.ChatRoom) error {
	// Add business logic/validation here if needed
	return s.repo.Create(chatroom)
}

func (s *ChatRoomService) GetChatRoomByID(id uint) (*models.ChatRoom, error) {
	chatroom, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatRoomNotFound
		}
		return nil, err
	}
	return chatroom, nil
}

func (s *ChatRoomService) ListChatRooms() ([]*models.ChatRoom, error) {
	return s.repo.List()
}

func (s *ChatRoomService) UpdateChatRoom(chatroom *models.ChatRoom) error {
	// Add business logic/validation here if needed
	return s.repo.Update(chatroom)
}

func (s *ChatRoomService) DeleteChatRoom(id uint) error {
	// Add business logic/validation here if needed
	return s.repo.Delete(id)
}

func (s *ChatRoomService) AddParticipantToChatRoom(chatroomID uint, userID uint) error {
	// Add business logic/validation here if needed
	return s.repo.AddParticipant(chatroomID, userID)
}

func (s *ChatRoomService) RemoveParticipantFromChatRoom(chatroomID uint, userID uint) error {
	// Add business logic/validation here if needed
	return s.repo.RemoveParticipant(chatroomID, userID)
}
