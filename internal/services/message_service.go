package services

import (
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

type MessageService struct {
	messageRepo repositories.MessageRepository
}

func NewMessageService(messageRepo *repositories.MessageRepository) *MessageService {
	return &MessageService{messageRepo: *messageRepo}
}

func (s *MessageService) CreateMessage(message *models.Message) error {
	return s.messageRepo.Create(message)
}

func (s *MessageService) GetMessagesByRoomID(
	roomID uint,
	limit, offset int,
) ([]models.Message, error) {
	return s.messageRepo.GetByRoomID(roomID, limit, offset)
}

func (s *MessageService) GetCountByRoomID(roomID uint) (int, error) {
	return s.messageRepo.GetCountByRoomID(roomID)
}
