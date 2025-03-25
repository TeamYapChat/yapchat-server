package models

import "gorm.io/gorm"

type ChatRoomType string

const (
	DirectMessageRoom ChatRoomType = "dm"
	GroupChatRoom     ChatRoomType = "group"
)

type ChatRoom struct {
	gorm.Model
	Name         string       `gorm:"varchar(255);index"`
	Type         ChatRoomType `gorm:"type:enum('dm', 'group');default:'dm'"`
	Participants []*User      `gorm:"many2many:chat_room_participants;"`
}

type ChatRoomRequest struct {
	Name           string       `json:"name,omitempty"`
	Type           ChatRoomType `json:"type"`
	ParticipantIDs []string     `json:"participant_ids,omitempty"`
}
