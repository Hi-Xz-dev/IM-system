package server

import (
	"testing"

	"IM-system/user"
)

func TestRenameSync(t *testing.T) {
	s := NewServer("127.0.0.1", 8080, nil)
	// 模拟 TCP连接
	tcpUser := &user.User{
		ID:          1,
		Nickname:    "Tom",
		Addr:        "127.0.0.1:10001",
		C:           make(chan string, 100),
		JoinedRooms: make(map[string]struct{}),
	}
	// 模拟 WebSocket连接
	wsUser := &user.User{
		ID:          1,
		Nickname:    "Tom",
		Addr:        "127.0.0.1:10002",
		C:           make(chan string, 100),
		JoinedRooms: make(map[string]struct{}),
	}

	s.Online(tcpUser)

	s.Online(wsUser)

	s.CreateRoom(tcpUser, "golang")

	s.Rename(tcpUser, "Jerry")

	// 全局名字应该更新
	if tcpUser.Nickname != "Jerry" {
		t.Fatalf("expected user Jerry online")
	}
	// 2. WebSocket昵称同步
	if wsUser.Nickname != "Jerry" {
		t.Fatalf(
			"expected ws user Jerry, got %s",
			wsUser.Nickname,
		)
	}
	users, ok := s.OnlineUsers[1]

	if !ok {
		t.Fatal("expected user id 1 online")
	}

	if len(users) != 2 {
		t.Fatalf(
			"expected 2 sessions, got %d",
			len(users),
		)
	}

	r := s.Rooms["golang"]
	roomUsers, ok := r.Users[1]

	if !ok {
		t.Fatal("expected user in room")
	}

	if roomUsers[0].Nickname != "Jerry" {
		t.Fatalf(
			"expected room user Jerry, got %s",
			roomUsers[0].Nickname,
		)
	}
}

