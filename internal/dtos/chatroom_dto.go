package dtos

import "github.com/teamyapchat/yapchat-server/internal/models"

type ChatRoomRequest struct {
	Name           string              `json:"name,omitempty"            example:"My Group Chat"`
	Type           models.ChatRoomType `json:"type"                      example:"group"`
	ParticipantIDs []string            `json:"participant_ids,omitempty"`
	ImageURL       string              `json:"image_url,omitempty"       example:"https://example.com/profile_picture.jpg"`
}

type ChatRoomResponse struct {
	ID           uint           `json:"id"`
	Name         string         `json:"name"`
	Type         string         `json:"type"`
	Participants []UserResponse `json:"participants"`
	ImageURL     string         `json:"image_url,omitempty"`
}
