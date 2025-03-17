package websocket

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type WSHandler struct {
	chatroomService *services.ChatRoomService
	messageService  *services.MessageService
	userService     *services.UserService
	clients         map[uint]*websocket.Conn
	nc              *nats.Conn
}

func NewWSHandler(
	natsURL string,
	chatroomService *services.ChatRoomService,
	messageService *services.MessageService,
	userService *services.UserService,
) *WSHandler {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS", "err", err.Error())
	}

	log.Info("Connected to NATS")

	return &WSHandler{
		chatroomService: chatroomService,
		messageService:  messageService,
		userService:     userService,
		clients:         make(map[uint]*websocket.Conn),
		nc:              nc,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin for now
	},
}

var mutex sync.Mutex

type Message struct {
	Content   string `json:"content"             binding:"required"`
	SenderID  uint   `json:"sender_id,omitempty"`
	RoomID    uint   `json:"room_id"             binding:"required"`
	Timestamp string `json:"timestamp,omitempty"`
	Type      string `json:"type,omitempty"`
}

// WebSocketHandler godoc
// @Summary      Handle websocket connection
// @Description  Handles websocket connections for real-time communication.
// @Tags         websocket
// @Router       /v1/ws [get]
func (h *WSHandler) WebSocketHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("user ID not found in context"),
		)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Failed to upgrade connection", "err", err.Error())
		return
	}
	defer func() {
		conn.Close()
		mutex.Lock()
		delete(h.clients, userID.(uint))
		mutex.Unlock()

		// Set status to offline
		_, err := h.userService.UpdateUser(
			userID.(uint),
			utils.UpdateUserRequest{Status: "offline"},
		)
		if err != nil {
			log.Error(
				"Failed to set user status to offline",
				"userID",
				userID.(uint),
				"err",
				err.Error(),
			)
		}

		log.Info("Client disconnected", "id", userID.(uint))
	}()

	mutex.Lock()
	h.clients[userID.(uint)] = conn
	mutex.Unlock()

	_, err = h.userService.UpdateUser(userID.(uint), utils.UpdateUserRequest{Status: "online"})
	if err != nil {
		log.Error(
			"Failed to set user status to online",
			"userID",
			userID.(uint),
			"err",
			err.Error(),
		)
	}

	log.Info("Client connected", "id", userID.(uint))

	// Handle panics in connection handler
	defer func() {
		if r := recover(); r != nil {
			log.Error("WebSocket panic recovered",
				"userID", userID.(uint),
				"panic", r,
				"stack", string(debug.Stack()))
		}
	}()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure) {
				log.Warn("Unexpected WebSocket closure",
					"userID", userID.(uint),
					"err", err.Error())
			}
			break
		}

		log.Debug("Received message", "msg", msg)

		msg.SenderID = userID.(uint)
		msg.Timestamp = time.Now().Format(time.RFC3339)
		msg.Type = "message"

		// Persist message to DB
		err = h.messageService.CreateMessage(&models.Message{
			SenderID:  msg.SenderID,
			RoomID:    msg.RoomID,
			Content:   msg.Content,
			Timestamp: time.Now(),
			Type:      msg.Type,
		})
		if err != nil {
			log.Error("Failed to persist message", "err", err.Error())
		}

		// Publish message to NATS
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			log.Error("Error marshaling message to JSON for NATS", "err", err.Error())
			continue // Handle error appropriately
		}

		err = h.nc.Publish("chat_messages", msgJSON)
		if err != nil {
			log.Error("Error publishing message to NATS", "err", err.Error())
			continue // Handle error appropriately
		}
	}
}

func (h *WSHandler) StartBroadcaster() {
	// Subscribe to NATS subject
	_, err := h.nc.Subscribe("chat_messages", func(m *nats.Msg) {
		var msg Message
		err := json.Unmarshal(m.Data, &msg)
		if err != nil {
			log.Error("Error unmarshaling NATS message", "err", err.Error())
			return
		}

		log.Debug("Received NATS message", "msg", msg)

		chatroom, err := h.chatroomService.GetChatRoomByID(msg.RoomID)
		if err != nil {
			log.Error("Error finding chatroom", "chatroomID", msg.RoomID, "err", err.Error())
			return
		}

		if chatroom.Participants == nil {
			log.Warn("No participants found in chatroom", "chatroomID", msg.RoomID)
			return
		}

		recipientIDs := make([]uint, 0, len(chatroom.Participants))
		for _, participant := range chatroom.Participants {
			recipientIDs = append(recipientIDs, participant.ID)
		}

		// Broadcast message to all connected clients, filtering by roomID
		mutex.Lock()
		defer mutex.Unlock()
		for _, userID := range recipientIDs {
			if client, exists := h.clients[userID]; exists {
				if err := client.WriteJSON(msg); err != nil {
					log.Error(
						"Error broadcasting message to client",
						"userID",
						userID,
						"err",
						err.Error(),
					)
					client.Close()
					delete(h.clients, userID) // Remove client if write fails
				}
			}
		}
	})
	if err != nil {
		log.Error("Error subscribing to NATS subject", "err", err.Error())
		return // Handle error appropriately
	}
	log.Info("Subscribed to NATS subject: chat_messages")

	// Keep broadcaster running
	select {} // Block indefinitely to keep goroutine alive
}
