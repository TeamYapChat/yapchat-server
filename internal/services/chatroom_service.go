package services

import (
	"errors"

	"github.com/charmbracelet/log"
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

var ErrChatRoomNotFound = errors.New("chat room not found")

type ChatRoomService struct {
	chatroomRepo *repositories.ChatRoomRepository
	userRepo     repositories.UserRepository
}

func NewChatRoomService(
	chatroomRepo *repositories.ChatRoomRepository,
	userRepo repositories.UserRepository,
) *ChatRoomService {
	return &ChatRoomService{chatroomRepo: chatroomRepo, userRepo: userRepo}
}

func (s *ChatRoomService) Create(chatroomReq *models.ChatRoomRequest) error {
	var participants []*models.User
	for _, id := range chatroomReq.ParticipantIDs {
		user, err := s.userRepo.FindByID(id)
		if err != nil {
			log.Warn(
				"Failed to find user while creating chat room",
				"userID",
				id,
				"err",
				err.Error(),
			)
			continue
		}

		participants = append(participants, user)
	}

	chatroom := models.ChatRoom{
		Name:         chatroomReq.Name,
		Type:         chatroomReq.Type,
		Participants: participants,
	}

	return s.chatroomRepo.Create(&chatroom)
}

func (s *ChatRoomService) GetByID(id uint) (*models.ChatRoom, error) {
	chatroom, err := s.chatroomRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatRoomNotFound
		}
		return nil, err
	}
	return chatroom, nil
}

func (s *ChatRoomService) List(userID string) ([]*models.ChatRoom, error) {
	return s.chatroomRepo.List(userID)
}

func (s *ChatRoomService) Update(chatroom *models.ChatRoom) error {
	// Add business logic/validation here if needed
	return s.chatroomRepo.Update(chatroom)
}

func (s *ChatRoomService) Delete(id uint) error {
	// Add business logic/validation here if needed
	return s.chatroomRepo.Delete(id)
}

func (s *ChatRoomService) AddParticipant(chatroomID uint, userID string) error {
	// Add business logic/validation here if needed
	return s.chatroomRepo.AddParticipant(chatroomID, userID)
}

func (s *ChatRoomService) RemoveParticipant(chatroomID uint, userID string) error {
	// Add business logic/validation here if needed
	return s.chatroomRepo.RemoveParticipant(chatroomID, userID)
}
