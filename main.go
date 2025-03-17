package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-contrib/graceful"
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
		&models.RefreshToken{},
	); err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(25)

	return db, nil
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

	db, err := InitDB(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database", "err", err.Error())
	}
	log.Info("Successfully initialized database")

	// Middlewares
	limiter := middleware.NewRateLimiter(cfg.RedisURL)

	limiter.AddLimiter("auth", middleware.RateLimitConfig{
		Limit:  5,
		Window: time.Minute,
	})

	limiter.AddLimiter("protected", middleware.RateLimitConfig{
		Limit:  5,
		Window: time.Second,
	})

	// Repos
	userRepo := repositories.NewUserRepository(db)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(db)
	chatroomRepo := repositories.NewChatRoomRepository(db)
	messageRepo := repositories.NewMessageRepository(db)

	// Services
	mailerService := services.NewMailerSendService(cfg.MailerSendAPIKey, cfg.EmailTemplateID)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, mailerService, cfg.JWTSecret)
	userService := services.NewUserService(userRepo)
	chatroomService := services.NewChatRoomService(chatroomRepo, userRepo)
	messageService := services.NewMessageService(messageRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	chatroomHandler := handlers.NewChatRoomHandler(chatroomService, messageService)
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

	public := router.Group("/auth")
	public.Use(limiter.Middleware("auth"))
	{
		public.GET("/verify-email", authHandler.VerifyEmailHandler)
		public.GET("/validate", authHandler.ValidateTokenHandler)

		public.POST("/register", authHandler.RegisterHandler)
		public.POST("/login", authHandler.LoginHandler)
		public.POST("/send-verification-email", authHandler.SendEmailHandler)
		public.POST("/refresh", authHandler.RefreshTokenHandler)
	}

	protected := router.Group("/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret), limiter.Middleware("protected"))
	{
		// User routes
		protected.GET("/user", userHandler.GetUser)
		protected.PUT("/user", userHandler.UpdateUser)
		protected.DELETE("/user", userHandler.DeleteUser)

		// Chatroom routes
		protected.GET("/chatrooms", chatroomHandler.ListChatRooms)
		protected.GET("/chatrooms/:id", chatroomHandler.GetChatRoomByID)
		protected.GET("/chatrooms/:id/messages", chatroomHandler.GetMessagesByRoomID)

		protected.POST("/chatrooms", chatroomHandler.CreateChatRoom)
		protected.POST("/chatrooms/:id/join", chatroomHandler.JoinChatRoom)
		protected.POST("/chatrooms/:id/leave", chatroomHandler.LeaveChatRoom)

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
