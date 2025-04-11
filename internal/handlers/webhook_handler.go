package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	svix "github.com/svix/svix-webhooks/go"

	"github.com/teamyapchat/yapchat-server/internal/dtos"
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type UserData struct {
	ID       string `json:"id"`
	Username string `json:"username,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type WebhookData struct {
	Data UserData `json:"data"`
	Type string   `json:"type"`
}

type WebhookHandler struct {
	webhook     *svix.Webhook
	userService *services.UserService
}

func NewWebhookHandler(signingSecret string, userService *services.UserService) *WebhookHandler {
	wh, err := svix.NewWebhook(signingSecret)
	if err != nil {
		log.Fatal("Failed to create Svix webhook", "err", err.Error())
	}

	return &WebhookHandler{
		webhook:     wh,
		userService: userService,
	}
}

func (h *WebhookHandler) WebhookHandler(c *gin.Context) {
	headers := c.Request.Header
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Failed to read payload"))
		return
	}

	if err := h.webhook.Verify(payload, headers); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Failed to verify payload"))
		return
	}

	var data WebhookData
	if err := json.Unmarshal(payload, &data); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid payload body"))
		return
	}

	if data.Type == "user.created" {
		if err := h.handleUserCreated(data.Data); err != nil {
			log.Error("Failed to create user", "userID", data.Data.ID, "err", err.Error())
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to create user"))
			return
		}
	} else if data.Type == "user.updated" {
		if err := h.handleUserUpdated(data.Data); err != nil {
			if err.Error() == "record not found" {
				log.Info("User not found, attempting to create user", "userID", data.Data.ID)
				err = h.handleUserCreated(data.Data)
			}
			if err != nil {
				log.Error("Failed to update user", "userID", data.Data.ID, "err", err.Error())
				c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to update user"))
				return
			}
		}
	} else if data.Type == "user.deleted" {
		if err := h.handleUserDeleted(data.Data); err != nil {
			log.Error("Failed to delete user", "userID", data.Data.ID, "err", err.Error())
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to delete user"))
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid event type"))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse{
		Success: true,
		Message: "success",
	})
}

// TODO: Integrate user service for database syncing
func (h *WebhookHandler) handleUserCreated(data UserData) error {
	return h.userService.Create(&models.User{
		ID:       data.ID,
		Username: data.Username,
		ImageURL: data.ImageURL,
	})
}

func (h *WebhookHandler) handleUserUpdated(data UserData) error {
	_, err := h.userService.Update(data.ID, dtos.UpdateUserRequest{
		Username: data.Username,
		ImageURL: data.ImageURL,
	})

	return err
}

func (h *WebhookHandler) handleUserDeleted(data UserData) error {
	return h.userService.Delete(data.ID)
}
