package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"

	"github.com/teamyapchat/yapchat-server/internal/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin for now
	},
}

var (
	clients = make(map[uint]*websocket.Conn)
	mutex   sync.Mutex
	nc      *nats.Conn
)

type Message struct {
	Content   string `json:"content"`
	SenderID  uint   `json:"sender_id"`
	RoomID    uint   `json:"room_id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
}

func InitializeNATS(url string) error {
	var err error
	nc, err = nats.Connect(url)
	if err != nil {
		return err
	}
	log.Info("Connected to NATS")
	return nil
}

// WebSocketHandler godoc
// @Summary      Handle websocket connection
// @Description  Handles websocket connections for real-time communication.
// @Tags         websocket
// @Router       /v1/ws [get]
func WebSocketHandler(c *gin.Context) {
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
	defer conn.Close()

	mutex.Lock()
	clients[userID.(uint)] = conn // Register new client with userID as key
	mutex.Unlock()
	defer func() {
		mutex.Lock()
		delete(clients, userID.(uint)) // Unregister client on disconnect using userID
		mutex.Unlock()
	}()

	log.Info("Client connected", "id", userID.(uint))

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Error("Error reading json message", "id", userID.(uint), "err", err.Error())
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

		err = nc.Publish("chat_messages", msgJSON)
		if err != nil {
			log.Error("Error publishing message to NATS", "err", err.Error())
			continue // Handle error appropriately
		}
	}
}

func StartBroadcaster() {
	// Subscribe to NATS subject
	_, err := nc.Subscribe("chat_messages", func(m *nats.Msg) {
		var msg Message
		err := json.Unmarshal(m.Data, &msg)
		if err != nil {
			log.Error("Error unmarshaling NATS message", "err", err.Error())
			return
		}

		log.Debug("Received NATS message", "msg", msg)

		// Broadcast message to all connected clients, filtering by roomID
		mutex.Lock()
		for userID, client := range clients { // Iterate through all connected clients
			// TODO: Retrieve user's room ID from message payload (msg.RoomID) and check if the connected client (userID) is in that room.
			// For now, just broadcast to all connected clients for testing.
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
				delete(clients, userID) // Remove client if write fails
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
