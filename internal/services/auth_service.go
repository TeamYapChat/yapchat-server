package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
)

type AuthService struct {
	UserRepo         repositories.UserRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	mailer           *MailerSendService
	jwtSecret        string
	baseURL          string
}

func NewAuthService(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	mailer *MailerSendService,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		UserRepo:         userRepo,
		RefreshTokenRepo: refreshTokenRepo,
		mailer:           mailer,
		jwtSecret:        jwtSecret,
		baseURL:          "https://yapchat.xyz",
	}
}

func (s *AuthService) Register(user *models.User) error {
	_, err := s.UserRepo.FindByEmail(user.Email)
	if err == nil {
		return errors.New("user already exists")
	}

	user.VerificationCode = generateVerificationCode()
	user.IsVerified = false

	if err := s.UserRepo.Create(user); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) Login(login, password string) (string, string, error) {
	var user *models.User
	var err error

	user, err = s.UserRepo.FindByEmail(login)
	if err != nil {
		user, err = s.UserRepo.FindByUsername(login)
		if err != nil {
			return "", "", errors.New("invalid credentials")
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}

	// Generate JWT access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(30 * time.Minute).Unix(),
	})
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshTokenValue := uuid.New().String()
	hashedRefreshToken, err := bcrypt.GenerateFromPassword(
		[]byte(refreshTokenValue),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", "", err
	}

	// Store refresh token in database
	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: string(hashedRefreshToken),
		Expiry:    time.Now().Add(7 * 24 * time.Hour),
	}
	err = s.RefreshTokenRepo.Create(refreshTokenModel)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenValue, nil
}

func (s *AuthService) RefreshToken(refreshTokenValue string) (string, string, error) {
	hashedRefreshToken, err := bcrypt.GenerateFromPassword(
		[]byte(refreshTokenValue),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	refreshTokenModel, err := s.RefreshTokenRepo.FindByTokenHash(string(hashedRefreshToken))
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if refreshTokenModel.Expiry.Before(time.Now()) {
		return "", "", errors.New("refresh token expired")
	}

	if refreshTokenModel.RevokedAt != nil {
		return "", "", errors.New("refresh token revoked")
	}

	user, err := s.UserRepo.FindByID(refreshTokenModel.UserID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// Generate new access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(30 * time.Minute).Unix(),
	})
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Generate new refresh token (rotation)
	newRefreshTokenValue := uuid.New().String()
	hashedNewRefreshToken, err := bcrypt.GenerateFromPassword(
		[]byte(newRefreshTokenValue),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", "", err
	}

	// Store new refresh token in database, replace old one
	refreshTokenModel.TokenHash = string(hashedNewRefreshToken)
	refreshTokenModel.Expiry = time.Now().Add(7 * 24 * time.Hour)
	refreshTokenModel.RevokedAt = nil
	err = s.RefreshTokenRepo.Revoke(refreshTokenModel)
	if err != nil {
		return "", "", err
	}
	refreshTokenModel.ID = 0
	err = s.RefreshTokenRepo.Create(refreshTokenModel)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, newRefreshTokenValue, nil
}

func (s *AuthService) SendVerificationEmail(id uint) error {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if user.IsVerified {
		return errors.New("user already verified")
	}

	verificationURL := s.baseURL + "/auth/verify?code=" + user.VerificationCode
	if err := s.mailer.SendVerificationEmail(user.Email, user.Username, verificationURL); err != nil {
		return errors.New("failed to send verification email")
	}

	return nil
}

func (s *AuthService) VerifyEmail(code string) error {
	user, err := s.UserRepo.FindByVerificationCode(code)
	if err != nil {
		return errors.New("user not found")
	}

	if user.IsVerified {
		return errors.New("user already verified")
	}

	user.IsVerified = true
	user.VerificationCode = ""

	if err := s.UserRepo.Update(user); err != nil {
		return err
	}

	return nil
}

func generateVerificationCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 64)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
