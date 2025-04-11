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
