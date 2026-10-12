package model

import (
	"time"
)

type Message struct {
	ID           int64
	Type         string
	FromID       int64
	FromNickname string
	ToID         *int64
	RoomID       *int64
	RoomName     *string
	Content      string
	CreatedAt    time.Time
}
