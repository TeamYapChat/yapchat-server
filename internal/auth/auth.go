package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	log "github.com/teamyapchat/yapchat-server/internal/logging"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(username string) (string, error) {
	secretKey := getJWTSecret()

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenString string) (*Claims, error) {
	secretKey := getJWTSecret()

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, err
	}

	return claims, nil
}

func getJWTSecret() string {
	secretPath := "/run/secrets/jwt-secret"

	data, err := os.ReadFile(secretPath)
	if err != nil {
		log.Error.Fatalln(
			"JWT Secret not found. Ensure the secret is correctly configured in Docker.",
		)
	}

	return string(data)
}
