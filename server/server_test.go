package server

import (
	"testing"

	"IM-system/user"
)

func TestOnlineOffline(t *testing.T) {
	s := NewServer("127.0.0.1", 8080, nil)

	u := &user.User{
		ID:       1,
		Nickname: "Tom",
		Addr:     "127.0.0.1:10001",
		C:        make(chan string, 100),
	}

	s.Online(u)

	users, ok := s.OnlineUsers[u.ID]

	if !ok {
		t.Fatalf("expected user online")
	}

	if len(users) != 1 {
		t.Fatalf(
			"expected 1 session, got %d",
			len(users),
		)
	}

	if users[0] != u {
		t.Fatalf("expected same user")
	}

	s.Offline(u)

	// 用户应该不存在
	users, ok = s.OnlineUsers[u.ID]

	if ok && len(users) != 0 {
		t.Fatalf(
			"expected user offline",
		)
	}

	if !u.IsClosed {
		t.Fatalf(
			"expected IsClosed=true",
		)
	}
}
func TestOfflineDoubleCall(t *testing.T) {
	s := NewServer("127.0.0.1", 8080, nil)

	u := &user.User{
		ID:          1,
		Nickname:    "Tom",
		Addr:        "127.0.0.1:10001",
		C:           make(chan string, 100),
		JoinedRooms: make(map[string]struct{}),
	}

	s.Online(u)

	s.Offline(u)

	users, ok := s.OnlineUsers[u.ID]

	if ok && len(users) != 0 {
		t.Fatalf(
			"expected user offline",
		)
	}

	if !u.IsClosed {
		t.Fatalf(
			"expected IsClosed=true",
		)
	}

	// 第二次调用
	s.Offline(u)

	// 再检查状态
	users, ok = s.OnlineUsers[u.ID]

	if ok && len(users) != 0 {
		t.Fatalf(
			"expected user still offline",
		)
	}

	if !u.IsClosed {
		t.Fatalf(
			"expected IsClosed=true after second call",
		)
	}
}
