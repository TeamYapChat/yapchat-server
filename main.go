package main

import (
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
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

	if err := db.AutoMigrate(&models.User{}); err != nil {
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
		log.Fatal("Failed to initialize database", "err", err)
	}
	log.Info("Successfully initialized database")

	userRepo := repositories.NewUserRepository(db)

	mailer := services.NewMailerSendService(cfg.MailerSendAPIKey, cfg.EmailTemplateID)

	router := gin.Default()
	router.Use(middleware.CORS())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := router.Group("/auth")
	{
		authService := services.NewAuthService(userRepo, mailer, cfg.JWTSecret)
		authHandler := handlers.NewAuthHandler(authService)

		public.GET("/verify-email", authHandler.VerifyEmailHandler)

		public.POST("/register", authHandler.RegisterHandler)
		public.POST("/login", authHandler.LoginHandler)
		public.POST("/send-verification-email", authHandler.SendEmailHandler)
	}

	protected := router.Group("/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		userService := services.NewUserService(userRepo)
		userHandler := handlers.NewUserHandler(userService)

		protected.GET("/user", userHandler.GetUser)
		protected.POST("/user", userHandler.UpdateUser)
		protected.DELETE("/user", userHandler.DeleteUser)
	}

	router.Run(":8080")
}
