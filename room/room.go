package room

import (
	"IM-system/user"
)
type Room struct {
	Name string
	Users  map[string]*user.User
	
}
func NewRoom(name string) *Room{
	return &Room{
		Name: name, 
		Users: make(map[string]*user.User),
	}
} 