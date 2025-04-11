package handlers

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/dtos"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetHandler godoc
// @Summary      Get user profile
// @Description  Get details of the currently authenticated user
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  utils.SuccessResponse{data=dtos.UserResponse}
// @Failure      401  {object}  utils.ErrorResponse
// @Failure      404  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /v1/user [get]
func (h *UserHandler) GetHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("User ID not found in context"),
		)
		return
	}

	user, err := h.userService.GetByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("User not found"))
		return
	}

	userResponse := dtos.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		ImageURL:  user.ImageURL,
		IsOnline:  user.IsOnline,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(userResponse))
}

// GetByUsernameHandler godoc
// @Summary      Get user profile by username
// @Description  Get details of a user using their username
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Param        username path string true "Username of the user to retrieve"
// @Success      200  {object}  utils.SuccessResponse{data=dtos.UserResponse}
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      401  {object}  utils.ErrorResponse
// @Failure      404  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /v1/user/{username} [get]
func (h *UserHandler) GetByUsernameHandler(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Username parameter not found"))
		return
	}

	user, err := h.userService.GetByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.NewErrorResponse("User not found"))
		log.Error("Failed to find user by username", "username", username, "err", err.Error())
		return
	}

	userResponse := dtos.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		ImageURL: user.ImageURL,
		IsOnline: user.IsOnline,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(userResponse))
}
