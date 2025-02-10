package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

type AuthService struct {
	userRepo  repositories.UserRepository
	mailer    *MailerSendService
	jwtSecret string
	baseURL   string
}

func NewAuthService(
	repo repositories.UserRepository,
	mailer *MailerSendService,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  repo,
		mailer:    mailer,
		jwtSecret: jwtSecret,
		baseURL:   "http://yapchat.xyz",
	}
}

func (s *AuthService) Register(user *models.User) error {
	_, err := s.userRepo.FindUserByEmail(user.Email)
	if err == nil {
		return errors.New("user already exists")
	}

	user.VerificationCode = generateVerificationCode()
	user.IsVerified = false

	if err := s.userRepo.CreateUser(user); err != nil {
		return err
	}

	verificationURL := s.baseURL + "/auth/verify-email?code=" + user.VerificationCode
	if err := s.mailer.SendVerificationEmail(user.Email, user.Username, verificationURL); err != nil {
		return errors.New("failed to send verification email")
	}

	return nil
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func generateVerificationCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
