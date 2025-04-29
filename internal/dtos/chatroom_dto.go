package dtos

import "github.com/teamyapchat/yapchat-server/internal/models"

type ChatRoomRequest struct {
	Name           string              `json:"name,omitempty"            example:"My Group Chat"`
	Type           models.ChatRoomType `json:"type"                      example:"group"                                   validate:"required"`
	ParticipantIDs []string            `json:"participant_ids,omitempty"`
	ImageURL       string              `json:"image_url,omitempty"       example:"https://example.com/profile_picture.jpg"`
}

type ChatRoomResponse struct {
	ID           uint           `json:"id"                  validate:"required"`
	Name         string         `json:"name"                validate:"required"`
	Type         string         `json:"type"                validate:"required"`
	Participants []UserResponse `json:"participants"        validate:"required"`
	ImageURL     string         `json:"image_url,omitempty"`
}
