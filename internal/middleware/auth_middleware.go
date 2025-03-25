package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/teamyapchat/yapchat-server/internal/utils"
)

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

func AuthMiddleware(store *JWKStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				utils.NewErrorResponse("Missing Authorization header"),
			)
			return
		}

		sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
		if sessionToken == authHeader {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				utils.NewErrorResponse("Invalid token format"),
			)
			return
		}

		unsafeClaims, err := jwt.Decode(c.Request.Context(), &jwt.DecodeParams{Token: sessionToken})
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				utils.NewErrorResponse("Invalid token"),
			)
			return
		}

		keyID := unsafeClaims.KeyID
		if keyID == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				utils.NewErrorResponse("Missing key ID in token"),
			)
			return
		}

		jwk := store.GetJWK(keyID)
		if jwk == nil {
			jwk, err = jwt.GetJSONWebKey(c.Request.Context(), &jwt.GetJSONWebKeyParams{
				KeyID:      keyID,
				JWKSClient: store.jwksClient,
			})
			if err != nil {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					utils.NewErrorResponse("Failed to fetch JWK"),
				)
				return
			}

			store.SetJWK(keyID, jwk)
		}

		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: sessionToken,
			JWK:   jwk,
		})
		if err != nil {
			log.Error("Failed to verify token", "err", err.Error())

			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				utils.NewErrorResponse("Invalid or expired token"),
			)
			return
		}

		usr, err := user.Get(c.Request.Context(), claims.Subject)
		if err != nil {
			log.Error("Failed to find user in Clerk", "id", claims.Subject, "err", err.Error())

			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				utils.NewErrorResponse("User not found"),
			)
			return
		}

		c.Set("userID", usr.ID)
		c.Next()
	}
}
