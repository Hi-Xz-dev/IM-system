package server

import (
	"testing"

	"IM-system/user"
)

func TestRenameSync(t *testing.T) {
	s := NewServer("127.0.0.1", 8080, nil)

	u := &user.User{
		Nickname: "Tom",
		Addr: "127.0.0.1:10001",
		C:    make(chan string, 100),
		JoinedRooms: make(map[string]struct{}),
	}

	s.Online(u)
	s.CreateRoom(u, "golang")
	s.Rename(u, "Jerry")

	// 全局名字应该更新
	if _, ok := s.OnlineUsers["Jerry"]; !ok {
		t.Fatalf("expected user Jerry online")
	}
	//旧名字应该被删除
	if _, ok := s.OnlineUsers["Tom"]; ok {
		t.Fatalf("expected old user Tom removed")
	}
	//房间内旧名字也应该删除
	r := s.Rooms["golang"]
	if _, ok := r.Users["Tom"]; ok {
		t.Fatalf("expected old user Tom removed from room")
	}
	// 当前房间用户名应该更新
	if _, ok := r.Users["Jerry"]; !ok {
		t.Fatalf("expected Jerry in room")

	}
}
