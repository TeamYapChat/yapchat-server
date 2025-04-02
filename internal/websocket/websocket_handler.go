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
	"github.com/mitchellh/mapstructure"
	"github.com/nats-io/nats.go"

	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/utils"
)

type WSHandler struct {
	authService     *services.AuthService
	chatroomService *services.ChatRoomService
	messageService  *services.MessageService
	userService     *services.UserService
	clients         map[string]*websocket.Conn
	nc              *nats.Conn
}

func NewWSHandler(
	natsURL string,
	authService *services.AuthService,
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
		authService:     authService,
		chatroomService: chatroomService,
		messageService:  messageService,
		userService:     userService,
		clients:         make(map[string]*websocket.Conn),
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
	Content   string `json:"content"`
	SenderID  string `json:"sender_id"`
	RoomID    uint   `json:"room_id"`
	Timestamp string `json:"timestamp"`
}

// WebSocketHandler godoc
// @Summary      Handle websocket connection
// @Description  Handles websocket connections for real-time communication.
// @Tags         websocket
// @Router       /ws [get]
func (h *WSHandler) WebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Failed to upgrade connection", "err", err.Error())
		return
	}

	var payload models.Payload
	if err := conn.ReadJSON(&payload); err != nil {
		if websocket.IsUnexpectedCloseError(err,
			websocket.CloseGoingAway,
			websocket.CloseNormalClosure) {
			log.Warn("Unexpected WebSocket closure",
				"err", err.Error())
		}
		conn.WriteJSON(gin.H{"error": "invalid payload structure"})
		conn.Close()

		log.Error("Failed to read message payload", "err", err.Error())
		return
	}

	if payload.Opcode != 1 {
		conn.WriteJSON(gin.H{"error": "invalid opcode"})
		conn.Close()

		log.Warn("Client did not send identify message")
		return
	}

	var identifyData models.IdentifyData
	if err := mapstructure.Decode(payload.Data, &identifyData); err != nil {
		conn.WriteJSON(gin.H{"error": "invalid data structure"})
		conn.Close()

		log.Error("Failed to unmarshal identify data", "err", err.Error())
		return
	}

	usr, err := h.authService.VerifyToken(c, identifyData.Token)
	if err != nil {
		conn.WriteJSON(gin.H{"error": err.Error()})
		conn.Close()

		return
	}

	userID := usr.ID

	defer func() {
		conn.Close()
		mutex.Lock()
		delete(h.clients, userID)
		mutex.Unlock()

		// Set status to offline
		_, err := h.userService.Update(
			userID,
			utils.UpdateUserRequest{Status: "offline"},
		)
		if err != nil {
			log.Error(
				"Failed to set user status to offline",
				"userID",
				userID,
				"err",
				err.Error(),
			)
		}

		log.Info("Client disconnected", "id", userID)
	}()

	mutex.Lock()
	h.clients[userID] = conn
	mutex.Unlock()

	_, err = h.userService.Update(userID, utils.UpdateUserRequest{Status: "online"})
	if err != nil {
		log.Error(
			"Failed to set user status to online",
			"userID",
			userID,
			"err",
			err.Error(),
		)
	}

	conn.WriteJSON(gin.H{"message": "successfully connected to server"})
	log.Info("Client connected", "id", userID)

	// Handle panics in connection handler
	defer func() {
		if r := recover(); r != nil {
			log.Error("WebSocket panic recovered",
				"userID", userID,
				"panic", r,
				"stack", string(debug.Stack()))
		}
	}()

	for {
		var payload models.Payload
		err := conn.ReadJSON(&payload)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure) {
				log.Warn("Unexpected WebSocket closure",
					"userID", userID,
					"err", err.Error())
			}
			break
		}

		if payload.Opcode != 0 {
			log.Warn("Invalid message type received", "opcode", payload.Opcode, "userID", userID)
			continue
		}

		log.Debug("Received message", "msg", payload)

		var msgData models.DispatchData
		if err := mapstructure.Decode(payload.Data, &msgData); err != nil {
			conn.WriteJSON(gin.H{"error": "invalid data body structure"})

			log.Error("Failed to unmarshal dispatch data", "err", err.Error())
			continue
		}

		msg := Message{
			Content:   msgData.Content,
			SenderID:  userID,
			RoomID:    msgData.RoomID,
			Timestamp: payload.Timestamp.Format(time.RFC3339),
		}

		// Persist message to DB
		err = h.messageService.CreateMessage(&models.Message{
			SenderID:  msg.SenderID,
			RoomID:    msg.RoomID,
			Content:   msg.Content,
			Timestamp: payload.Timestamp,
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

		chatroom, err := h.chatroomService.GetByID(msg.RoomID)
		if err != nil {
			log.Error("Error finding chatroom", "chatroomID", msg.RoomID, "err", err.Error())
			return
		}

		if chatroom.Participants == nil {
			log.Warn("No participants found in chatroom", "chatroomID", msg.RoomID)
			return
		}

		recipientIDs := make([]string, 0, len(chatroom.Participants))
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
