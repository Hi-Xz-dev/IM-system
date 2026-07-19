package server

import (
	"testing"

	"IM-system/user"
)

func TestRoomJoinLeave(t *testing.T) {
	s := NewServer("127.0.0.1", 8080, nil)

	u := &user.User{
		Nickname:        "Tom",
		Addr:        "127.0.0.1:10001",
		C:           make(chan string, 100),
		JoinedRooms: make(map[string]struct{}),
	}

	s.Online(u)
	s.CreateRoom(u, "golang")
	// 创建房间后，用户应该进入该房间
	if _, ok := u.JoinedRooms["golang"]; !ok {
		t.Fatalf("expected current room golang, got %s", "golang")
	}
	// 房间应该已经创建
	r, ok := s.Rooms["golang"]
	if !ok {
		t.Fatalf("expected room golang exists")
	}
	// 房间成员应该包含 Tom
	if _, ok := r.Users["Tom"]; !ok {
		t.Fatalf("expected Tom in room")
	}
	s.LeaveRoom(u, "golang")

	// 离开房间后，用户当前房间应该为空
	if _, ok := u.JoinedRooms["golang"]; ok {
		t.Fatalf("expected current room empty, got %s", "golang")
	}

	// 房间为空后，应该自动删除
	if _, ok := s.Rooms["golang"]; ok {
		t.Fatalf("expected room golang deleted")
	}
}
