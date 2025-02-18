package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Response Structs
type UserResponse struct {
	ID       uint   `json:"id"       example:"123"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email"    example:"john@example.com"`
}

// GetUser godoc
// @Summary      Get user profile
// @Description  Get details of the currently authenticated user
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Success      200  {object}  utils.SuccessResponse{data=UserResponse}
// @Failure      401  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /v1/user [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("user ID not found in context"),
		)
		return
	}

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(err.Error()))
		return
	}

	userResponse := UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(userResponse))
}

// UpdateUser godoc
// @Summary      Update user profile
// @Description  Update details of the currently authenticated user
// @Tags         users
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Param        request body models.User true "User details to update"
// @Success      200  {object}  utils.SuccessResponse{data=UserResponse}
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      401  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /v1/user [post]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("user ID not found in context"),
		)
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse(err.Error()))
		return
	}

	updatedUser, err := h.userService.UpdateUser(userID.(string), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(err.Error()))
		return
	}

	userResponse := UserResponse{
		ID:       updatedUser.ID,
		Username: updatedUser.Username,
		Email:    updatedUser.Email,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(userResponse))
}

// DeleteUser godoc
// @Summary      Delete user profile
// @Description  Delete the currently authenticated user's profile
// @Tags         users
// @Security     ApiKeyAuth
// @Produce      json
// @Success      204  "No Content"
// @Failure      401  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /v1/user [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("user ID not found in context"),
		)
		return
	}

	err := h.userService.DeleteUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
