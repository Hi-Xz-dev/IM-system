package server

import (
	"testing"
	"time"

	"IM-system/user"
)

func TestListenDisconnect(t *testing.T) {
	u := &user.User{
		ID:       1,
		Nickname: "Tom",
	}

	s := &Server{
		OnlineUsers: make(map[int64][]*user.User),
		Disconnect:  make(chan *user.User, 1),
	}

	s.OnlineUsers[u.ID] = []*user.User{u}

	go s.ListenDisconnect()

	s.Disconnect <- u

	time.Sleep(100 * time.Millisecond)
	
	s.mapLock.RLock()
	
	if _, ok := s.OnlineUsers[u.ID]; ok {

		s.mapLock.RUnlock()

		t.Fatal("user still online")
	}
}
