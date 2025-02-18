package websocket

import (
	"net/http"
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
	clients   = make(map[*websocket.Conn]bool) // connected clients
	broadcast = make(chan Message)             // broadcast channel
	mutex     sync.Mutex                       // to protect clients map
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

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Failed to upgrade connection", "err", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true // Register new client
	mutex.Unlock()
	defer func() {
		mutex.Lock()
		delete(clients, conn) // Unregister client when connection closes
		mutex.Unlock()
	}()

	log.Info("Client connected", "id", userID.(uint))

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Error("Error reading json message", "id", userID.(uint), "err", err)
			break
		}

		log.Debug("Received message", "msg", msg)

		broadcast <- msg
	}
}

func StartBroadcaster() {
	for {
		msg := <-broadcast

		mutex.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Error("Error broadcasting message to client", "err", err)
				client.Close()
				mutex.Unlock()
				delete(clients, client)
				mutex.Lock()
			}
		}
		mutex.Unlock()
	}
}

func init() {
	go StartBroadcaster()
}
