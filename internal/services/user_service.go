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
	return s.userRepo.FindUserByID(id)
}

func (s *UserService) UpdateUser(id uint, data utils.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindUserByID(id)
	if err != nil {
		return nil, err
	}

	if data.Username != "" {
		user.Username = data.Username
	}
	if data.ImageURL != "" {
		user.ImageURL = data.ImageURL
	}

	err = s.userRepo.UpdateUser(user)
	return user, err
}

func (s *UserService) DeleteUser(id uint) error {
	return s.userRepo.DeleteUser(id)
}
