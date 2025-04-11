package services

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/teamyapchat/yapchat-server/internal/dtos"
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

var ErrChatRoomNotFound = errors.New("chat room not found")

type ChatRoomService struct {
	chatroomRepo *repositories.ChatRoomRepository
	userRepo     repositories.UserRepository
	rdb          *redis.Client
}

func NewChatRoomService(
	chatroomRepo *repositories.ChatRoomRepository,
	userRepo repositories.UserRepository,
	redisClient *redis.Client,
) *ChatRoomService {
	return &ChatRoomService{
		chatroomRepo: chatroomRepo,
		userRepo:     userRepo,
		rdb:          redisClient,
	}
}

func (s *ChatRoomService) Create(chatroomReq *dtos.ChatRoomRequest) error {
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

func (s *ChatRoomService) CreateInviteCode(chatroomID uint) (string, error) {
	for {
		key := generateULID()

		exists, err := s.rdb.Exists(context.Background(), "invite:"+key).Result()
		if err != nil {
			log.Error("Failed to check if invite code exists", "err", err.Error())
			return "", err
		}

		if exists == 0 {
			err = s.rdb.Set(context.Background(), "invite:"+key, chatroomID, 24*time.Hour).Err()
			if err != nil {
				log.Error("Failed to set invite code", "err", err.Error())
				return "", err
			}

			return key, nil
		}
	}
}

func (s *ChatRoomService) GetByInviteCode(inviteCode string) (*models.ChatRoom, error) {
	chatroomIDStr, err := s.rdb.Get(context.Background(), "invite:"+inviteCode).Result()
	if err != nil {
		log.Error("Failed to get invite code", "err", err.Error())
		return nil, err
	}

	chatroomID, err := strconv.ParseUint(chatroomIDStr, 10, 32)
	if err != nil {
		log.Error("Failed to parse chatroom ID", "err", err.Error())
		return nil, err
	}

	chatroom, err := s.GetByID(uint(chatroomID))
	if err != nil {
		log.Error("Failed to get chat room", "err", err.Error())
		return nil, err
	}

	return chatroom, nil
}

func generateULID() string {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
