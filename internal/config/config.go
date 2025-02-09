package config

import (
	"os"

	"github.com/charmbracelet/log"
)

type Config struct {
	DBHost           string
	DBUser           string
	DBPassword       string
	DBName           string
	JWTSecret        string
	MailerSendAPIKey string
	EmailTemplateID  string
}

func LoadConfig() (config Config) {
	config = Config{
		DBHost:           os.Getenv("DB_HOST"),
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       getSecret("DB_PASS_FILE"),
		DBName:           os.Getenv("DB_NAME"),
		JWTSecret:        getSecret("JWT_SECRET_FILE"),
		MailerSendAPIKey: getSecret("MAILERSEND_API_KEY_FILE"),
		EmailTemplateID:  os.Getenv("EMAIL_TEMPLATE_ID"),
	}

	return
}

func getSecret(envName string) string {
	secretPath := os.Getenv(envName)

	data, err := os.ReadFile(secretPath)
	if err != nil {
		log.Fatal("Environment variable not set", "env", secretPath, "err", err)
	}

	return string(data)
}
