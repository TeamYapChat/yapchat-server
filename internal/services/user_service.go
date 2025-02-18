package services

import (
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetUserByID(id string) (*models.User, error) {
	uid := stringToUint(id)

	return s.userRepo.FindUserByID(uid)
}

func (s *UserService) UpdateUser(id string, user models.User) (*models.User, error) {
	uid := stringToUint(id)
	user.ID = uint(uid)
	err := s.userRepo.UpdateUser(&user)

	return &user, err
}

func (s *UserService) DeleteUser(id string) error {
	uid := stringToUint(id)
	uID := uint(uid)

	return s.userRepo.DeleteUser(uID)
}

func stringToUint(s string) uint {
	n := uint(0)
	for _, c := range s {
		if '0' <= c && c <= '9' {
			n = n*10 + uint(c-'0')
		}
	}

	return n
}
