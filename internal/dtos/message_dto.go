package dtos

import "time"

type MessageResponse struct {
	Content   string `json:"content"           validate:"required"`
	SenderID  string `json:"sender_id"         validate:"required"`
	RoomID    uint   `json:"room_id,omitempty"`
	Timestamp string `json:"timestamp"         validate:"required"`
}

type Payload struct {
	Opcode    int            `json:"op"`
	Data      map[string]any `json:"data"`
	Timestamp time.Time      `json:"timestamp"`
}

// Opcode 0
type DispatchData struct {
	Content string `mapstructure:"content"`
	RoomID  uint   `mapstructure:"room_id"`
}

// Opcode 1
type IdentifyData struct {
	Token string `mapstructure:"token"`
}
