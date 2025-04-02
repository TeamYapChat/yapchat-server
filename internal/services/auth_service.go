package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/charmbracelet/log"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	store *JWKStore
}

func NewAuthService(store *JWKStore) *AuthService {
	return &AuthService{
		store: store,
	}
}

func (s *AuthService) VerifyToken(ctx context.Context, sessionToken string) (*clerk.User, error) {
	unsafeClaims, err := jwt.Decode(ctx, &jwt.DecodeParams{Token: sessionToken})
	if err != nil {
		log.Error("Failed to decode JWT", "err", err.Error())
		return nil, errors.New("invalid token")
	}

	keyID := unsafeClaims.KeyID
	if keyID == "" {
		log.Error("Failed to extract key ID from claims")
		return nil, errors.New("invalid token")
	}

	jwk := s.store.GetJWK(keyID)
	if jwk == nil {
		jwk, err = jwt.GetJSONWebKey(ctx, &jwt.GetJSONWebKeyParams{
			KeyID:      keyID,
			JWKSClient: s.store.jwksClient,
		})
		if err != nil {
			log.Error("Failed to fetch JWK", "err", err.Error())
			return nil, errors.New("failed to fetch JWK")
		}

		s.store.SetJWK(keyID, jwk)
	}

	claims, err := jwt.Verify(ctx, &jwt.VerifyParams{
		Token: sessionToken,
		JWK:   jwk,
	})
	if err != nil {
		log.Error("Failed to verify token", "err", err.Error())
		return nil, errors.New("invalid or expired token")
	}

	usr, err := user.Get(ctx, claims.Subject)
	if err != nil {
		log.Error("Failed to find user in Clerk", "id", claims.Subject, "err", err.Error())
		return nil, errors.New("user not found")
	}

	return usr, nil
}

type JWKStore struct {
	jwksClient  *jwks.Client
	redisClient *redis.Client
}

func NewJWKStore(clerkSecretKey string, redisClient *redis.Client) *JWKStore {
	config := &clerk.ClientConfig{}
	config.Key = clerk.String(clerkSecretKey)

	return &JWKStore{
		jwksClient:  jwks.NewClient(config),
		redisClient: redisClient,
	}
}

func (s *JWKStore) GetJWK(keyID string) *clerk.JSONWebKey {
	ctx := context.Background()
	jwkData, err := s.redisClient.Get(ctx, "jwks:"+keyID).Bytes()
	if err != nil {
		return nil
	}

	var jwk clerk.JSONWebKey
	if err := json.Unmarshal(jwkData, &jwk); err != nil {
		return nil
	}

	return &jwk
}

func (s *JWKStore) SetJWK(keyID string, jwk *clerk.JSONWebKey) {
	ctx := context.Background()
	jwkData, err := json.Marshal(jwk)
	if err != nil {
		log.Error("Failed to serialize JWK", "err", err.Error())
		return
	}

	err = s.redisClient.Set(ctx, "jwks:"+keyID, jwkData, time.Hour).Err()
	if err != nil {
		log.Error("Failed to cache JWK", "err", err.Error())
	}
}
