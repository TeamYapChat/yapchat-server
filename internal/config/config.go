package config

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string // "prod", "dev", "test"
	DBHost        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisURL      string
	NATSURL       string
	ClerkSecret   string
	SigningSecret string
}

func LoadConfig() Config {
	var config Config

	config.AppEnv = strings.ToLower(os.Getenv("APP_ENV"))
	if config.AppEnv == "" {
		config.AppEnv = "test" // Default to test
	}

	if config.AppEnv == "dev" || config.AppEnv == "test" {
		log.SetLevel(log.DebugLevel)

		if err := godotenv.Load(); err != nil {
			log.Warn("No .env file found. Using environment variables directly.")
		}
	}

	config.DBHost = os.Getenv("DB_HOST")
	config.DBUser = os.Getenv("DB_USER")
	config.DBName = os.Getenv("DB_NAME")
	config.RedisURL = os.Getenv("REDIS_URL")
	config.NATSURL = os.Getenv("NATS_URL")
	config.DBPassword = os.Getenv("DB_PASS")
	config.ClerkSecret = os.Getenv("CLERK_SECRET_KEY")
	config.SigningSecret = os.Getenv("SIGNING_SECRET")

	return config
}
