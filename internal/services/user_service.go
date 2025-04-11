package services

import (
	"github.com/teamyapchat/yapchat-server/internal/dtos"
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Create(user *models.User) error {
	return s.userRepo.Create(user)
}

func (s *UserService) GetByID(id string) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) GetByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) Update(id string, data dtos.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if data.Username != "" {
		user.Username = data.Username
		err = s.userRepo.UpdateUsername(user)
	}
	if data.ImageURL != "" {
		user.ImageURL = data.ImageURL
		err = s.userRepo.UpdateImage(user)
	}
	if data.Status != "" {
		user.IsOnline = data.Status == "online"
		err = s.userRepo.UpdateStatus(user)
	}

	return user, err
}

func (s *UserService) Delete(id string) error {
	return s.userRepo.Delete(id)
}
