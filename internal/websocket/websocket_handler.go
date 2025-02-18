package websocket

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

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
	clients   = make(map[uint]map[*websocket.Conn]bool) // connected clients per room
	broadcast = make(chan struct {
		RoomID  uint
		Message Message
	}) // broadcast channel with RoomID
	mutex sync.Mutex // to protect clients map
)

type Message struct {
	Content string `json:"content"`
}

// WebSocketHandler godoc
// @Summary      Handle websocket connection
// @Description  Handles websocket connections for real-time communication.
// @Tags         websocket
// @Router       /ws [get]
func WebSocketHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(
			http.StatusInternalServerError,
			utils.NewErrorResponse("user ID not found in context"),
		)
		return
	}

	roomIDStr := c.Query("room_id")
	if roomIDStr == "" {
		c.JSON(
			http.StatusBadRequest,
			utils.NewErrorResponse("room_id query parameter is required"),
		)
		return
	}

	roomIDUint64, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		c.JSON(
			http.StatusBadRequest,
			utils.NewErrorResponse("invalid room_id format"),
		)
		return
	}
	roomID := uint(roomIDUint64)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Failed to upgrade connection", "err", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	if _, ok := clients[roomID]; !ok {
		clients[roomID] = make(map[*websocket.Conn]bool)
	}
	clients[roomID][conn] = true // Register new client for room
	mutex.Unlock()
	defer func() {
		mutex.Lock()
		delete(clients[roomID], conn) // Unregister client when connection closes
		if len(clients[roomID]) == 0 {
			delete(clients, roomID) // Remove room if no clients left
		}
		mutex.Unlock()
	}()

	log.Info("Client connected", "id", userID.(uint), "roomID", roomID)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Error("Error reading json message", "id", userID.(uint), "err", err)
			break
		}

		log.Debug("Received message", "msg", msg)

		broadcast <- struct {
			RoomID  uint
			Message Message
		}{
			RoomID:  roomID,
			Message: msg,
		}
	}
}

func StartBroadcaster() {
	for {
		broadcastMsg := <-broadcast

		mutex.Lock()
		if roomClients, ok := clients[broadcastMsg.RoomID]; ok {
			for client := range roomClients {
				err := client.WriteJSON(broadcastMsg.Message)
				if err != nil {
					log.Error("Error broadcasting message to client", "err", err)
					client.Close()
					delete(roomClients, client) // Remove client from roomClients map
				}
			}
		}
		mutex.Unlock()
	}
}

func init() {
	go StartBroadcaster()
}
