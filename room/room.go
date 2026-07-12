package room

import (
	"IM-system/user"
)

type Room struct {
	Name  string
	//当前房间成员
	Users map[string]*user.User
}

func NewRoom(name string) *Room {
	return &Room{
		Name:  name,
		Users: make(map[string]*user.User),
	}
}
