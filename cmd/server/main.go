package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/teamyapchat/yapchat-server/internal/auth"
	"github.com/teamyapchat/yapchat-server/internal/database"
	"github.com/teamyapchat/yapchat-server/internal/handlers"
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

	r.Static("/", "./dist")

	r.POST("/login", handlers.Login)
	api := r.Group("/api")
	api.Use(AuthMiddleware())
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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("username", claims.Username)
		c.Next()
	}
}
