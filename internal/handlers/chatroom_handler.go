package handlers

import (
	"errors"
	"math"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/dtos"
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

// CreateHandler godoc
// @Summary      Create a new chat room
// @Description  Create a new chat room
// @Tags         chatrooms
// @Accept       json
// @Produce      json
// @Param        request body dtos.ChatRoomRequest true "Chat room info"
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

	var chatroomRequest dtos.ChatRoomRequest
	if err := c.ShouldBindJSON(&chatroomRequest); err != nil {
		c.JSON(http.StatusBadRequest, utils.NewErrorResponse("Invalid request body"))
		return
	}

	chatroomRequest.ParticipantIDs = append(chatroomRequest.ParticipantIDs, userID.(string))

	if err := h.chatroomService.Create(&chatroomRequest); err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to create chat room"))
		return
	}

	c.JSON(http.StatusCreated,
		utils.SuccessResponse{
			Success: true,
			Message: "Chat room created successfully",
		},
	)
}

// GetByIDHandler godoc
// @Summary      Get chat room by ID
// @Description  Get chat room by ID
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Success      200 {object} utils.SuccessResponse{data=dtos.ChatRoomResponse}
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

	inviteCode := c.Query("code")
	if inviteCode != "" {
		inviteRoom, err := h.chatroomService.GetByInviteCode(inviteCode)
		if err != nil {
			if errors.Is(err, services.ErrChatRoomNotFound) {
				c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
			} else {
				c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to get chat room"))
			}
			return
		}
		if inviteRoom.ID != id {
			c.JSON(
				http.StatusBadRequest,
				utils.NewErrorResponse("Invalid invite code for this chat room"),
			)
			return
		}
	}

	if !slices.Contains(getParticipantIDs(chatroom.Participants), userID.(string)) &&
		inviteCode == "" {
		c.JSON(http.StatusForbidden, utils.NewErrorResponse("User not in chat room"))
		return
	}

	response := dtos.ChatRoomResponse{
		ID:           chatroom.ID,
		Name:         chatroom.Name,
		Type:         string(chatroom.Type),
		Participants: getParticipants(chatroom.Participants),
		ImageURL:     chatroom.ImageURL,
	}

	c.JSON(http.StatusOK, utils.NewSuccessResponse(response))
}

// GetMessagesByRoomIDHandler godoc
// @Summary      Get messages by chat room ID
// @Description  Get messages for a specific chat room
// @Tags         chatrooms
// @Produce      json
// @Param        id path integer true "Chat room ID"
// @Param        page query integer false "Page number (default 1)"
// @Param        page_size query integer false "Number of messages per page (default 25)"
// @Success      200 {object} utils.Pagination{data=[]dtos.MessageResponse}
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

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 1
	}
	if page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil {
		pageSize = 25
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	totalRows, err := h.messageService.GetCountByRoomID(id)
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("Failed to get message count"),
		)
		return
	}
	totalPages := int(math.Ceil(float64(totalRows) / float64(pageSize)))

	if page > totalPages || totalRows == 0 {
		pagination := utils.Pagination{
			Page:       page,
			PageSize:   pageSize,
			TotalRows:  totalRows,
			TotalPages: totalPages,
			Data:       []dtos.MessageResponse{},
		}

		c.JSON(http.StatusOK, pagination)
		return
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

	messages, err := h.messageService.GetMessagesByRoomID(chatroom.ID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to get messages"))
		return
	}

	messageList := make([]dtos.MessageResponse, 0, len(messages))
	for _, message := range messages {
		messageList = append(messageList, dtos.MessageResponse{
			Content:   message.Content,
			SenderID:  message.SenderID,
			Timestamp: message.Timestamp.Format(time.RFC3339),
		})
	}

	pagination := utils.Pagination{
		Page:       page,
		PageSize:   pageSize,
		TotalRows:  totalRows,
		TotalPages: totalPages,
		Data:       messageList,
	}

	c.JSON(http.StatusOK, pagination)
}

// ListChatroomsHandler godoc
// @Summary      List all chat rooms
// @Description  Get a list of all chat rooms that the user is in
// @Tags         chatrooms
// @Produce      json
// @Success      200 {object} utils.SuccessResponse{data=[]dtos.ChatRoomResponse}
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

	var responses []dtos.ChatRoomResponse
	for _, chatroom := range chatrooms {
		participants := getParticipants(chatroom.Participants)

		responses = append(responses, dtos.ChatRoomResponse{
			ID:           chatroom.ID,
			Name:         chatroom.Name,
			Type:         string(chatroom.Type),
			Participants: participants,
			ImageURL:     chatroom.ImageURL,
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
// @Failure      409 {object} utils.ErrorResponse
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

	if slices.Contains(getParticipantIDs(chatroom.Participants), userID.(string)) {
		c.JSON(http.StatusConflict, utils.NewErrorResponse("User already in chat room"))
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
// @Success      204
// @Failure      400 {object} utils.ErrorResponse
// @Failure      401 {object} utils.ErrorResponse
// @Failure      404 {object} utils.ErrorResponse
// @Failure      409 {object} utils.ErrorResponse
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
	chatroom, err := h.chatroomService.GetByID(chatroomID)
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to leave chat room"))
		}
		return
	}

	if !slices.Contains(getParticipantIDs(chatroom.Participants), userID.(string)) {
		c.JSON(http.StatusConflict, utils.NewErrorResponse("User not in chat room"))
		return
	}

	err = h.chatroomService.RemoveParticipant(chatroomID, userID.(string))
	if err != nil {
		if errors.Is(err, services.ErrChatRoomNotFound) {
			c.JSON(http.StatusNotFound, utils.NewErrorResponse("Chat room not found"))
		} else {
			c.JSON(http.StatusInternalServerError, utils.NewErrorResponse("Failed to leave chat room"))
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func getParticipants(participants []*models.User) []dtos.UserResponse {
	users := make([]dtos.UserResponse, 0, len(participants))
	for _, p := range participants {
		users = append(users, dtos.UserResponse{
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
