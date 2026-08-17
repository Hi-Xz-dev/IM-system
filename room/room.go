package room

import (
	"IM-system/user"
)

type Room struct {
	ID   int64
	Name string
	//当前房间成员 同一用户多端连接
	Users map[int64][]*user.User
}

func NewRoom(id int64, name string) *Room {
	return &Room{
		ID:    id,
		Name:  name,
		Users: make(map[int64][]*user.User),
	}
}
