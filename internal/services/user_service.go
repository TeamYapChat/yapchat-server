package services

import (
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) UpdateUser(id uint, data utils.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	user.Password = ""

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

func (s *UserService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}
