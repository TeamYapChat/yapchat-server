package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type MailerSendService struct {
	apiKey     string
	apiURL     string
	fromEmail  string
	fromName   string
	templateID string
}

type MailerSendRequest struct {
	From            EmailInfo                 `json:"from"`
	To              []EmailInfo               `json:"to"`
	Subject         string                    `json:"subject"`
	Personalization []TemplatePersonalization `json:"personalization"`
	TemplateID      string                    `json:"template_id"`
}

type EmailInfo struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type TemplatePersonalization struct {
	Email string       `json:"email"`
	Data  TemplateData `json:"data"`
}

type TemplateData struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

func NewMailerSendService(apiKey, templateID string) *MailerSendService {
	return &MailerSendService{
		apiKey:     apiKey,
		apiURL:     "https://api.mailersend.com/v1/email",
		fromEmail:  "no-reply@yapchat.xyz",
		fromName:   "YapChat",
		templateID: templateID,
	}
}

func (m *MailerSendService) SendVerificationEmail(email, name, verificationURL string) error {
	reqBody := MailerSendRequest{
		From: EmailInfo{
			Email: m.fromEmail,
			Name:  m.fromName,
		},
		To: []EmailInfo{
			{
				Email: email,
				Name:  name,
			},
		},
		Subject: "Confirm your email for your YapChat account",
		Personalization: []TemplatePersonalization{
			{
				Email: email,
				Data: TemplateData{
					Name: name,
					Link: verificationURL,
				},
			},
		},
		TemplateID: m.templateID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", m.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("failed to send verification email")
	}

	return nil
}
