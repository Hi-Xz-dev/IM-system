package room

import (
	"IM-system/user"
)

type Room struct {
	Name string
	//当前房间成员
	Users map[int64][]*user.User
}

func NewRoom(name string) *Room {
	return &Room{
		Name:  name,
		Users: make(map[int64][]*user.User),
	}
}
