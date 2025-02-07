package utils

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mailersend/mailersend-go"

	log "github.com/teamyapchat/yapchat-server/internal/logging"
	"github.com/teamyapchat/yapchat-server/internal/models"
)

func SendVerificationEmail(user models.User) error {
	apiKey := getAPIKey()

	ms := mailersend.NewMailersend(apiKey)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	verificationLink := fmt.Sprintf("http://yapchat.xyz/verify?token=%s", user.VerificationToken)

	subject := "Confirm your email for your YapChat account"

	from := mailersend.From{
		Name:  "YapChat",
		Email: "no-reply@yapchat.xyz",
	}

	recipient := []mailersend.Recipient{
		{
			Email: user.Email,
		},
	}

	personalization := []mailersend.Personalization{
		{
			Email: user.Email,
			Data: map[string]interface{}{
				"name": user.Username,
				"link": verificationLink,
			},
		},
	}

	templateID := os.Getenv("EMAIL_TEMPLATE_ID")

	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipient)
	message.SetSubject(subject)
	message.SetTemplateID(templateID)
	message.SetPersonalization(personalization)

	res, err := ms.Email.Send(ctx, message)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusAccepted {
		log.Warning.Println("Unexpected status code while sending email to:", user.Email)
	}

	log.Info.Println("Sent verification email to:", user.Email)
	return nil
}

func getAPIKey() string {
	secretPath := os.Getenv("MAILERSEND_API_KEY_FILE")

	data, err := os.ReadFile(secretPath)
	if err != nil {
		log.Error.Fatalln("MailerSend API Key not found.")
	}

	return string(data)
}
