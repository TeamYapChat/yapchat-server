package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessTokenCookie, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "User is not authorized"},
			)
			return
		}

		tokenString := accessTokenCookie
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "Invalid token"},
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"error": "Invalid token claims"},
			)
			return
		}

		c.Set("userID", uint(claims["sub"].(float64)))
		c.Next()
	}
}
