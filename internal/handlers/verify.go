package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	db "github.com/teamyapchat/yapchat-server/internal/database"
)

func Verify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	user, err := db.GetUserByVerificationToken(token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid token"})
		return
	}

	if err := db.VerifyUser(user.ID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}
