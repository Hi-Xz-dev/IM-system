package domain

import (
	"time"
)

type MessageType string

const (
	MessageText    MessageType = "text"
	MessagePrivate MessageType = "private"
	MessageRoom    MessageType = "room"
	MessageSystem  MessageType = "system"
)

type Message struct {
	ID           string      `json:"id"`
	Type         MessageType `json:"type"`
	From         int64       `json:"from"`
	FromNickname string      `json:"from_nickname"`
	To           int64       `json:"to,omitempty"`
	RoomID       int64       `json:"room_id,omitempty"`
	RoomName     string      `json:"room_name,omitempty"`
	Content      string      `json:"content"`
	Time         time.Time   `json:"time"`
}

func NewMessage(msgType MessageType, from int64, fromNickname, content string) Message {
	return Message{
		Type:         msgType,
		From:         from,
		FromNickname: fromNickname,
		Content:      content,
		Time:         time.Now(),
	}
}

// SystemMessage 系统消息（From=0, FromNickname="系统"）
func SystemMessage(content string) Message {
	return NewMessage(MessageSystem, 0, "系统", content)
}
