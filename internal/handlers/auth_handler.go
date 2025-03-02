package handlers

import (
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Request Structs
type RegisterRequest struct {
	Username string `json:"username" binding:"required"       example:"john_doe"`
	Email    string `json:"email"    binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

type LoginRequest struct {
	Login    string `json:"login"    binding:"required" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type SendEmailRequest struct {
	Id uint `json:"id" binding:"required" example:"123"`
}

type VerifyEmailRequest struct {
	Code string `form:"code" binding:"required"`
}

const domain string = ".yapchat.xyz"

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
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
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

	if err := h.authService.Register(&user); err != nil {
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
				err.Error(),
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

// LoginHandler godoc
// @Summary      Authenticate user
// @Description  Login with email or username and password. Access and refresh tokens are returned as HttpOnly cookies.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body LoginRequest true "User credentials"
// @Success      200 {object} utils.SuccessResponse "Successful login. Access and refresh tokens are in HttpOnly cookies."
// @Failure      400  {object}  utils.ErrorResponse
// @Failure      401  {object}  utils.ErrorResponse
// @Failure      500  {object}  utils.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			utils.ErrorResponse{Success: false, Message: "Invalid request body"},
		)
		return
	}

	accessToken, refreshToken, err := h.authService.Login(req.Login, req.Password)
	if err != nil {
		c.JSON(
			http.StatusUnauthorized,
			utils.ErrorResponse{Success: false, Message: "Invalid credentials"},
		)
		return
	}

	c.SetSameSite(http.SameSiteNoneMode)

	// Set access token cookie
	c.SetCookie(
		"access_token",
		accessToken,
		1800, // 30 minutes
		"/",
		domain,
		true, // Secure
		true, // HttpOnly
	)

	// Set refresh token cookie
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*3600, // 7 days
		"/",
		domain,
		true, // Secure
		true, // HttpOnly
	)

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Success: true,
		Message: "Login successful",
	})
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
func (h *AuthHandler) SendEmailHandler(c *gin.Context) {
	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			utils.ErrorResponse{Success: false, Message: "Invalid request body"},
		)
		return
	}

	if err := h.authService.SendVerificationEmail(req.Id); err != nil {
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
func (h *AuthHandler) VerifyEmailHandler(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			utils.ErrorResponse{Success: false, Message: "Invalid verification code"},
		)
		return
	}

	if err := h.authService.VerifyEmail(req.Code); err != nil {
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

// RefreshTokenHandler godoc
// @Summary      Refresh access and refresh tokens
// @Description  Handles refresh token logic to issue new access and refresh tokens. New access and refresh tokens are returned as HttpOnly cookies.
// @Tags         auth
// @Produce      json
// @Success      200 {object} utils.SuccessResponse "Successful token refresh. New access and refresh tokens are in HttpOnly cookies."
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	refreshTokenCookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Success: false,
			Message: "Refresh token cookie not found",
		})
		return
	}

	accessTokenString, newRefreshTokenValue, err := h.authService.RefreshToken(
		userID.(uint),
		refreshTokenCookie,
	)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "invalid refresh token" || err.Error() == "refresh token expired" ||
			err.Error() == "refresh token revoked" {
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, utils.ErrorResponse{
			Success: false,
			Message: "Failed to refresh tokens: " + err.Error(),
		})
		return
	}

	c.SetSameSite(http.SameSiteNoneMode)

	// Set new access token cookie
	c.SetCookie(
		"access_token",
		accessTokenString,
		1800, // 30 minutes
		"/",
		domain,
		true, // Secure
		true, // HttpOnly
	)

	// Set new refresh token cookie if rotation occurred
	if newRefreshTokenValue != "" {
		c.SetCookie(
			"refresh_token",
			newRefreshTokenValue,
			7*24*3600, // 7 days
			"/",
			domain,
			true, // Secure
			true, // HttpOnly
		)
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Success: true,
		Message: "Tokens refreshed successfully",
	})
}

// ValidateTokenHandler godoc
// @Summary      Validate access token
// @Description  Validates the access token from the cookie.
// @Tags         auth
// @Produce      json
// @Success      200 {object} utils.SuccessResponse
// @Failure      401 {object} utils.ErrorResponse
// @Router       /auth/validate [get]
func (h *AuthHandler) ValidateTokenHandler(c *gin.Context) {
	accessTokenCookie, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("Access token cookie not found"))
		return
	}

	if !h.authService.ValidateAccessToken(accessTokenCookie) {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("Invalid access token"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Success: true,
		Message: "Token is valid",
	})
}
