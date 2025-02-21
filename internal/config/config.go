package config

import (
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv           string // "prod", "dev", "test"
	DBHost           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTSecret        string
	MailerSendAPIKey string
	EmailTemplateID  string
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
	config.JWTSecret = os.Getenv("JWT_SECRET")
	config.EmailTemplateID = os.Getenv("EMAIL_TEMPLATE_ID")

	if config.AppEnv == "prod" {
		config.DBPassword = getSecret("DB_PASS_FILE")
		config.MailerSendAPIKey = getSecret("MAILERSEND_API_KEY_FILE")
	} else {
		config.DBPassword = os.Getenv("DB_PASS")
		config.MailerSendAPIKey = os.Getenv("MAILERSEND_API_KEY")
	}

	return config
}

func getSecret(envName string) string {
	secretPath := os.Getenv(envName)
	if secretPath == "" {
		log.Fatal("Environment variable for secret file path not set", "env", envName)
	}

	data, err := os.ReadFile(secretPath)
	if err != nil {
		log.Fatal(
			"Failed to read secret file",
			"path",
			secretPath,
			"env",
			envName,
			"err",
			err.Error(),
		)
	}

	return strings.TrimSpace(string(data))
}
