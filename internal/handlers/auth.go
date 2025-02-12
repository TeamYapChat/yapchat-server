package handlers

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

// Request Structs
type RegisterRequest struct {
	Username string `json:"username" binding:"required"       example:"john_doe"`
	Email    string `json:"email"    binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required"       example:"password123"`
}

type SendEmailRequest struct {
	Id uint `json:"id" binding:"required" example:"123"`
}

type VerifyEmailRequest struct {
	Code string `form:"code" binding:"required"`
}

// Response Structs
type UserResponse struct {
	ID       uint   `json:"id"       example:"123"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email"    example:"john@example.com"`
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

			if err.Error() == "failed to send verification email" {
				log.Error(
					"Failed to send verification email",
					"username",
					req.Username,
					"email",
					req.Email,
					"err",
					err,
				)
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

// SendEmailHandler godoc
// @Summary      Send verification email
// @Description  Send verification email to the user's email address
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body SendEmailRequest true "User ID"
// @Success      200  {object}  utils.SuccessResponse
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      404  {object}  utils.ErrorResponse
// @Failure      409  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /auth/send-verification-email [post]
func SendEmailHandler(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(
				http.StatusBadRequest,
				utils.ErrorResponse{Success: false, Message: "Invalid request body"},
			)
			return
		}

		if err := authService.SendVerificationEmail(req.Id); err != nil {
			if err.Error() == "user not found" {
				c.JSON(
					http.StatusNotFound,
					utils.ErrorResponse{Success: false, Message: "User not found"},
				)
				return
			}

			if err.Error() == "user already verified" {
				c.JSON(
					http.StatusBadRequest,
					utils.ErrorResponse{Success: false, Message: "Email already verified"},
				)
				return
			}

			c.JSON(
				http.StatusInternalServerError,
				utils.ErrorResponse{Success: false, Message: "Failed to send verification email"},
			)
			return
		}

		c.JSON(http.StatusOK, utils.SuccessResponse{
			Success: true,
			Message: "Verification email sent successfully",
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
func VerifyEmailHandler(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req VerifyEmailRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(
				http.StatusBadRequest,
				utils.ErrorResponse{Success: false, Message: "Invalid verification code"},
			)
			return
		}

		if err := authService.VerifyEmail(req.Code); err != nil {
			if err.Error() == "user not found" {
				c.JSON(
					http.StatusNotFound,
					utils.ErrorResponse{Success: false, Message: "Invalid verification code"},
				)
				return
			}

			if err.Error() == "user already verified" {
				c.JSON(
					http.StatusBadRequest,
					utils.ErrorResponse{Success: false, Message: "Email already verified"},
				)
				return
			}

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
