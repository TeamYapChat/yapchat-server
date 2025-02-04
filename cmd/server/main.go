package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/database"
	log "github.com/teamyapchat/yapchat-server/internal/logging"
)

func init() {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user,
		pass,
		host,
		dbName,
	)

	_, err := database.Connect(dsn)
	if err != nil {
		panic(err)
	}
	log.Info.Println("Successfully connected to database")
}

func main() {
	// DEBUG
	log.Debug.Println(GetOutboundIP())

	r := gin.Default()
	r.LoadHTMLFiles("web/home.html")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	r.Run("0.0.0.0:8080")
}

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Error.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
