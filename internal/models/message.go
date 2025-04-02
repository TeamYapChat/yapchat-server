package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	SenderID  string `gorm:"varchar(255);index"`
	Sender    User   `gorm:"foreignKey:SenderID"`
	RoomID    uint
	Room      ChatRoom `gorm:"foreignKey:RoomID"`
	Content   string
	Timestamp time.Time
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
