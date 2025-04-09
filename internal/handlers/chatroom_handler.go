package handlers

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type ChatRoomHandler struct {
	chatroomService *services.ChatRoomService
	messageService  *services.MessageService
}

func NewChatRoomHandler(
	chatroomService *services.ChatRoomService,
	messageService *services.MessageService,
) *ChatRoomHandler {
	return &ChatRoomHandler{
		chatroomService: chatroomService,
		messageService:  messageService,
	}
}

// ChatRoomResponse defines the response structure for chat room related API calls
type ChatRoomResponse struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Participants []UserResponse `json:"participants"`
}

type MessageResponse struct {
	Content   string `json:"content"`
	SenderID  string `json:"sender_id"`
	Timestamp string `json:"timestamp"`
}

// CreateHandler godoc
// @Summary      Create a new chat room
// @Description  Create a new chat room
// @Tags         chatrooms
// @Accept       json
// @Produce      json
// @Param        request body models.ChatRoomRequest true "Chat room info"
// @Success      201 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms [post]
func (h *ChatRoomHandler) CreateHandler(c *gin.Context) {
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

	chatroomRequest.ParticipantIDs = append(chatroomRequest.ParticipantIDs, userID.(string))

	if err := h.chatroomService.Create(&chatroomRequest); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to create chat room"))
		return
	}

	c.JSON(http.StatusCreated, utils.NewSuccessResponse("Chat room created successfully"))
}

// GetByIDHandler godoc
// @Summary      Get chat room by ID
// @Description  Get chat room by ID
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Success      200 {object} utils.SuccessResponse{data=ChatRoomResponse}
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms/{id} [get]
func (h *ChatRoomHandler) GetByIDHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid chat room ID"))
		return
	}
	id := uint(idUint64)

	chatroom, err := h.chatroomService.GetByID(id)
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to get chat room"))
		}
		return
	}

	if !slices.Contains(getParticipantIDs(chatroom.Participants), userID.(string)) {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("User not in chat room"))
		return
	}

	response := ChatRoomResponse{
		ID:           chatroom.ID,
		Name:         chatroom.Name,
		Type:         string(chatroom.Type),
		Participants: getParticipants(chatroom.Participants),
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(response))
}

// GetMessagesByRoomIDHandler godoc
// @Summary      Get messages by chat room ID
// @Description  Get messages for a specific chat room
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Param        count query integer false "Number of messages to return (default 25)"
// @Success      200 {object} utils.SuccessResponse{data=[]MessageResponse}
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      403 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms/{id}/messages [get]
func (h *ChatRoomHandler) GetMessagesByRoomIDHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid chat room ID"))
		return
	}
	id := uint(idUint64)

	count, err := strconv.Atoi(c.Query("count"))
	if err != nil {
		count = 25
	}

	chatroom, err := h.chatroomService.GetByID(id)
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to get chat room"))
		}
		return
	}

	if !slices.Contains(getParticipantIDs(chatroom.Participants), userID.(string)) {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("User not in chat room"))
		return
	}

	messages, err := h.messageService.GetMessagesByRoomID(chatroom.ID, count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to get messages"))
		return
	}

	messageList := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		messageList = append(messageList, MessageResponse{
			Content:   message.Content,
			SenderID:  message.SenderID,
			Timestamp: message.Timestamp.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(messageList))
}

// ListChatroomsHandler godoc
// @Summary      List all chat rooms
// @Description  Get a list of all chat rooms that the user is in
// @Tags         chatrooms
// @Produce      json
// @Success      200 {object} utils.SuccessResponse{data=[]ChatRoomResponse}
// @Failure      401 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms [get]
func (h *ChatRoomHandler) ListChatroomsHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	chatrooms, err := h.chatroomService.List(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to list chat rooms"))
		return
	}

	var responses []ChatRoomResponse
	for _, chatroom := range chatrooms {
		participants := getParticipants(chatroom.Participants)

		responses = append(responses, ChatRoomResponse{
			ID:           chatroom.ID,
			Name:         chatroom.Name,
			Type:         string(chatroom.Type),
			Participants: participants,
		})
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(responses))
}

// GetInviteCodeHandler godoc
// @Summary      Get an invite code for a chat room
// @Description  Create and return an invite code for a chat room
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Success      200 {object} utils.SuccessResponse{data=string}
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      403 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms/{id}/invite-code [get]
func (h *ChatRoomHandler) GetInviteCodeHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid chat room ID"))
		return
	}
	chatroomID := uint(idUint64)

	chatroom, err := h.chatroomService.GetByID(chatroomID)
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to get chat room"))
		}
		return
	}

	if !slices.Contains(getParticipantIDs(chatroom.Participants), userID.(string)) {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("User not in chat room"))
		return
	}

	inviteCode, err := h.chatroomService.CreateInviteCode(chatroomID)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("Failed to create invite code"),
		)
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(inviteCode))
}

// JoinChatroomHandler godoc
// @Summary      Join chat room by ID
// @Description  Join chat room by ID
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms/{id}/join [post]
func (h *ChatRoomHandler) JoinChatroomHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid chat room ID"))
		return
	}
	chatroomID := uint(idUint64)

	inviteCode := c.Query("code")
	if inviteCode == "" {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invite code is required"))
		return
	}

	chatroom, err := h.chatroomService.GetByInviteCode(inviteCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid invite code"))
		return
	}

	if chatroom.ID != chatroomID {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid invite code"))
		return
	}

	err = h.chatroomService.AddParticipant(chatroomID, userID.(string))
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to join chat room"))
		}
		return
	}

	c.JSON(http.StatusOK,
		utils.SuccessResponse{
			Success: true,
			Message: "Successfully joined chat room",
		},
	)
}

// LeaveChatroomHandler godoc
// @Summary      Leave chat room by ID
// @Description  Leave chat room by ID
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Success      200 {object} utils.SuccessResponse
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      500 {object} utils.ErrorResponse
// @Router       /v1/chatrooms/{id}/leave [post]
func (h *ChatRoomHandler) LeaveChatroomHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.NewErrorResponse("User ID not found in context"))
		return
	}

	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid chat room ID"))
		return
	}
	chatroomID := uint(idUint64)

	// TODO: Only remove user that exists in chat room

	err = h.chatroomService.RemoveParticipant(chatroomID, userID.(string))
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to leave chat room"))
		}
		return
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse("Successfully left chat room"))
}

func getParticipants(participants []*models.User) []UserResponse {
	users := make([]UserResponse, 0, len(participants))
	for _, p := range participants {
		users = append(users, UserResponse{
			ID:       p.ID,
			Username: p.Username,
			ImageURL: p.ImageURL,
			IsOnline: p.IsOnline,
		})
	}
	return users
}

func getParticipantIDs(participants []*models.User) []string {
	ids := make([]string, 0, len(participants))
	for _, p := range participants {
		ids = append(ids, p.ID)
	}
	return ids
}
