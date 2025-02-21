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
	repo     *repositories.ChatRoomRepository
	userRepo repositories.UserRepository
}

func NewChatRoomService(
	repo *repositories.ChatRoomRepository,
	userRepo repositories.UserRepository,
) *ChatRoomService {
	return &ChatRoomService{repo: repo, userRepo: userRepo}
}

func (s *ChatRoomService) CreateChatRoom(chatroomReq *models.ChatRoomRequest) error {
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

	return s.repo.Create(&chatroom)
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

func (s *ChatRoomService) ListChatRooms(userID uint) ([]*models.ChatRoom, error) {
	return s.repo.List(userID)
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
