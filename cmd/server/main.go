package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/database"
	"github.com/teamyapchat/yapchat-server/internal/handlers"
	log "github.com/teamyapchat/yapchat-server/internal/logging"
	"github.com/teamyapchat/yapchat-server/internal/middleware"
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

	if err := database.Connect(dsn); err != nil {
		log.Error.Fatalln("Failed to connect to database:", err)
	}
	log.Info.Println("Successfully connected to database.")

	if err := database.Migrate(); err != nil {
		log.Error.Fatalln("Failed to migrate database:", err)
	}
	log.Info.Println("Successfully migrated database.")
}

func main() {
	r := gin.Default()

	r.Use(middleware.RateLimitMiddleware())

	r.StaticFS("/assets", http.Dir("./dist"))

	r.GET("/verify", handlers.Verify)
	r.POST("/login", handlers.Login)
	r.POST("/register", handlers.Register)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/protected", func(c *gin.Context) {
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{"message": "Welcome!", "user": username})
		})
	}

	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	r.Run(":8081")
}
