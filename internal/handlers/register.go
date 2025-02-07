package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	db "github.com/teamyapchat/yapchat-server/internal/database"
)

func Register(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := db.GetUserByEmail(input.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Account already exists"})
		return
	}

	if _, err := db.GetUserByUsername(input.Username); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	if err := db.CreateUser(input.Email, input.Username, input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}
