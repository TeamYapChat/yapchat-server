package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type ChatRoomHandler struct {
	service *services.ChatRoomService
}

func NewChatRoomHandler(service *services.ChatRoomService) *ChatRoomHandler {
	return &ChatRoomHandler{service: service}
}

// CreateChatRoom godoc
// @Summary      Create a new chat room
// @Description  Create a new chat room
// @Tags         chatrooms
// @Accept       json
// @Produce      json
// @Param        request body models.ChatRoomRequest true "Chat room info"
// @Success      201 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms [post]
func (h *ChatRoomHandler) CreateChatRoom(c *gin.Context) {
	var chatroomRequest models.ChatRoomRequest
	if err := c.ShouldBindJSON(&chatroomRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid request body"))
		return
	}

	chatroom := models.ChatRoom{
		Name: chatroomRequest.Name,
		Type: chatroomRequest.Type,
	}

	if err := h.service.CreateChatRoom(&chatroom); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to create chat room"))
		return
	}

	c.JSON(http.StatusCreated, utils.NewSuccessResponse("Chat room created successfully"))
}
