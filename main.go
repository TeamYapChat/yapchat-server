package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-contrib/graceful"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/teamyapchat/yapchat-server/docs"
	"github.com/teamyapchat/yapchat-server/internal/config"
	"github.com/teamyapchat/yapchat-server/internal/handlers"
	"github.com/teamyapchat/yapchat-server/internal/middleware"
	"github.com/teamyapchat/yapchat-server/internal/models"
	"github.com/teamyapchat/yapchat-server/internal/repositories"
	"github.com/teamyapchat/yapchat-server/internal/services"
	"github.com/teamyapchat/yapchat-server/internal/websocket"
)

func InitDB(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.ChatRoom{},
		&models.Message{},
	); err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(25)

	return db, nil
}

func InitRedis(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(opts), nil
}

// @title           YapChat API
// @version         1.0
// @description     The official API for YapChat
//
// @license.name    GPLv3
// @license.url     https://www.gnu.org/licenses/gpl-3.0.en.html
//
// @host            api.yapchat.xyz
func main() {
	cfg := config.LoadConfig()

	// Initialize databases
	db, err := InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database", "err", err.Error())
	}
	log.Info("Successfully initialized database")

	redisClient, err := InitRedis(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL", "url", cfg.RedisURL, "err", err.Error())
	}
	log.Info("Successfully connected to Redis")

	// Middlewares
	limiter := middleware.NewRateLimiter(redisClient)

	limiter.AddLimiter("protected", middleware.RateLimitConfig{
		Limit:  5,
		Window: time.Second,
	})

	clerk.SetKey(cfg.ClerkSecret)
	store := middleware.NewJWKStore(cfg.ClerkSecret, redisClient)

	// Repos
	userRepo := repositories.NewUserRepository(db)
	chatroomRepo := repositories.NewChatRoomRepository(db)
	messageRepo := repositories.NewMessageRepository(db)

	// Services
	userService := services.NewUserService(userRepo)
	chatroomService := services.NewChatRoomService(chatroomRepo, userRepo)
	messageService := services.NewMessageService(messageRepo)

	// Handlers
	userHandler := handlers.NewUserHandler(userService)
	chatroomHandler := handlers.NewChatRoomHandler(chatroomService, messageService)
	webhookHandler := handlers.NewWebhookHandler(cfg.SigningSecret, userService)
	wsHandler := websocket.NewWSHandler(cfg.NATSURL, chatroomService, messageService, userService)
	go wsHandler.StartBroadcaster()

	router, err := graceful.Default()
	if err != nil {
		log.Fatal("Failed to create router", "err", err.Error())
	}
	defer router.Close()

	router.SetTrustedProxies(nil)
	router.Use(middleware.CORS())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/webhook", webhookHandler.WebhookHandler)

	protected := router.Group("/v1")
	protected.Use(middleware.AuthMiddleware(store), limiter.Middleware("protected"))
	{
		// User routes
		protected.GET("/user", userHandler.GetHandler)
		protected.GET("/user/:username", userHandler.GetByUsernameHandler)

		// Chatroom routes
		protected.GET("/chatrooms", chatroomHandler.ListChatroomsHandler)
		protected.GET("/chatrooms/:id", chatroomHandler.GetByIDHandler)
		protected.GET("/chatrooms/:id/messages", chatroomHandler.GetMessagesByRoomIDHandler)

		protected.POST("/chatrooms", chatroomHandler.CreateHandler)
		protected.POST("/chatrooms/:id/join", chatroomHandler.JoinChatroomHandler)
		protected.POST("/chatrooms/:id/leave", chatroomHandler.LeaveChatroomHandler)

		// Websocket routes
		protected.GET("/ws", wsHandler.WebSocketHandler)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run on :8080 by default
	if err := router.RunWithContext(ctx); err != nil && err == context.Canceled {
		log.Fatal("Shutting down server", "err", err.Error())
	}
}
