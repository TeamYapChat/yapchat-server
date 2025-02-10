package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"       example:"john_doe"`
	Email    string `json:"email"    binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required"       example:"password123"`
}

type UserResponse struct {
	ID       uint   `json:"id"       example:"123"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email"    example:"john@example.com"`
}

type VerifyEmailRequest struct {
	Code string `form:"code" binding:"required"`
}

// RegisterHandler godoc
// @Summary      Register new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body RegisterRequest true "User registration information"
// @Success      201  {object}  utils.SuccessResponse{data=UserResponse}
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      409  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /auth/register [post]
func RegisterHandler(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(
				http.StatusBadRequest,
				utils.ErrorResponse{Success: false, Message: "Invalid request body"},
			)
			return
		}

		user := models.User{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
		}

		if err := authService.Register(&user); err != nil {
			statusCode := http.StatusInternalServerError
			if err.Error() == "user already exists" {
				statusCode = http.StatusConflict
			}

			c.JSON(statusCode, utils.ErrorResponse{Success: false, Message: err.Error()})
			return
		}

		userResponse := UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		}

		c.JSON(http.StatusCreated, utils.SuccessResponse{
			Success: true,
			Message: "User registered successfully",
			Data:    userResponse,
		})
	}
}

// LoginHandler godoc
// @Summary      Authenticate user
// @Description  Login with email and password to receive JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body LoginRequest true "User credentials"
// @Success      200  {object}  utils.SuccessResponse{data=string}
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      401  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /auth/login [post]
func LoginHandler(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(
				http.StatusBadRequest,
				utils.ErrorResponse{Success: false, Message: "Invalid request body"},
			)
			return
		}

		token, err := authService.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(
				http.StatusUnauthorized,
				utils.ErrorResponse{Success: false, Message: "Invalid credentials"},
			)
			return
		}

		c.JSON(http.StatusOK, utils.SuccessResponse{
			Success: true,
			Message: "Login successful",
			Data:    token,
		})
	}
}

// VerifyEmailHandler godoc
// @Summary      Verify email address
// @Description  Verify user's email address using verification code
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        code query string true "Verification code"
// @Success      200  {object}  utils.SuccessResponse
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      404  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /auth/verify-email [get]
func VerifyEmailHandler(userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyEmailRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(
				http.StatusBadRequest,
				utils.ErrorResponse{Success: false, Message: "Invalid verification code"},
			)
			return
		}

		user, err := userRepo.FindUserByVerificationCode(req.Code)
		if err != nil {
			c.JSON(
				http.StatusNotFound,
				utils.ErrorResponse{Success: false, Message: "Invalid verification code"},
			)
			return
		}

		if user.IsVerified {
			c.JSON(
				http.StatusBadRequest,
				utils.ErrorResponse{Success: false, Message: "Email already verified"},
			)
			return
		}

		user.IsVerified = true
		user.VerificationCode = ""

		if err := userRepo.UpdateUser(user); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				utils.ErrorResponse{Success: false, Message: "Failed to verify email"},
			)
			return
		}

		c.JSON(http.StatusOK, utils.SuccessResponse{
			Success: true,
			Message: "Email verified successfully",
		})
	}
}
