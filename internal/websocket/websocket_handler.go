package websocket

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"slices"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"

	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type WSHandler struct {
	chatroomService *services.ChatRoomService
	clients         map[uint]*websocket.Conn
	nc              *nats.Conn
}

func NewWSHandler(natsURL string, chatroomService *services.ChatRoomService) *WSHandler {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("Failed to connect to NATS", "err", err.Error())
	}

	log.Info("Connected to NATS")

	return &WSHandler{
		chatroomService: chatroomService,
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
		log.Info("Client disconnected", "id", userID.(uint))
	}()

	mutex.Lock()
	h.clients[userID.(uint)] = conn
	mutex.Unlock()

	log.Info("Client connected", "id", userID.(uint))

	// Handle panics in connection handler
	defer func() {
		if r := recover(); r != nil {
			log.Error("WebSocket panic recovered",
				"id", userID.(uint),
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
					"id", userID.(uint),
					"err", err.Error())
			}
			break
		}

		log.Debug("Received message", "msg", msg)

		msg.SenderID = userID.(uint)
		msg.Timestamp = time.Now().Format(time.RFC3339)
		msg.Type = "message"

		// TODO: Persist message to DB

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

		var recipientIDs []uint
		if chatroom.Participants != nil {
			for _, participant := range chatroom.Participants {
				recipientIDs = append(recipientIDs, participant.ID)
			}
		}

		// Broadcast message to all connected clients, filtering by roomID
		mutex.Lock()
		for userID, client := range h.clients { // Iterate through all connected clients
			if !slices.Contains(recipientIDs, userID) {
				continue
			}

			err := client.WriteJSON(msg) // Use msg from NATS message
			if err != nil {
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
		mutex.Unlock()
	})
	if err != nil {
		log.Error("Error subscribing to NATS subject", "err", err.Error())
		return // Handle error appropriately
	}
	log.Info("Subscribed to NATS subject: chat_messages")

	// Keep broadcaster running
	select {} // Block indefinitely to keep goroutine alive
}
