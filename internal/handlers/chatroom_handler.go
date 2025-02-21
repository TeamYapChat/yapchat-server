package handlers

import (
	"errors"
	"net/http"
	"strconv"

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

// ChatRoomResponse defines the response structure for chat room related API calls
type ChatRoomResponse struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	ParticipantIDs []uint `json:"participant_ids"`
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
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	var chatroomRequest models.ChatRoomRequest
	if err := c.ShouldBindJSON(&chatroomRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid request body"))
		return
	}

	chatroomRequest.ParticipantIDs = append(chatroomRequest.ParticipantIDs, userID.(uint))

	if err := h.service.CreateChatRoom(&chatroomRequest); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to create chat room"))
		return
	}

	c.JSON(http.StatusCreated, utils.NewSuccessResponse("Chat room created successfully"))
}

// GetChatRoomByID godoc
// @Summary      Get chat room by ID
// @Description  Get chat room by ID
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Success      200 {object} ChatRoomResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms/{id} [get]
func (h *ChatRoomHandler) GetChatRoomByID(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid chat room ID"))
		return
	}
	id := uint(idUint64)

	chatroom, err := h.service.GetChatRoomByID(id)
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to get chat room"))
		}
		return
	}

	var userIDList []uint
	if chatroom.Participants != nil {
		for _, participant := range chatroom.Participants {
			userIDList = append(userIDList, participant.ID)
		}
	}

	response := ChatRoomResponse{
		ID:             chatroom.ID,
		Name:           chatroom.Name,
		Type:           string(chatroom.Type),
		ParticipantIDs: userIDList,
	}

	c.JSON(http.StatusOK, response)
}

// ListChatRooms godoc
// @Summary      List all chat rooms
// @Description  Get a list of all chat rooms
// @Tags         chatrooms
// @Produce      json
// @Success      200 {array} ChatRoomResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms [get]
func (h *ChatRoomHandler) ListChatRooms(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	chatrooms, err := h.service.ListChatRooms(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to list chat rooms"))
		return
	}

	var responses []ChatRoomResponse
	for _, chatroom := range chatrooms {
		var userIDList []uint
		if chatroom.Participants != nil {
			for _, participant := range chatroom.Participants {
				userIDList = append(userIDList, participant.ID)
			}
		}

		responses = append(responses, ChatRoomResponse{
			ID:             chatroom.ID,
			Name:           chatroom.Name,
			Type:           string(chatroom.Type),
			ParticipantIDs: userIDList,
		})
	}

	c.JSON(http.StatusOK, responses)
}
