package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	SenderID  uint      `json:"sender_id"`
	Sender    User      `gorm:"foreignKey:SenderID"`
	RoomID    uint      `json:"room_id"`
	Room      ChatRoom  `gorm:"foreignKey:RoomID"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
}
