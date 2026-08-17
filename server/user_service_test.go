package server

import (
	"testing"

	"IM-system/user"
	"IM-system/room"
)

func TestRenameSync(t *testing.T) {
	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1

	tcpUser := &user.User{
		ID:          1,
		Nickname:    "Tom",
		Addr:        "127.0.0.1:10001",
		C:           make(chan string, 100),
		JoinedRooms: make(map[int64]struct{}),
	}

	wsUser := &user.User{
		ID:          1,
		Nickname:    "Tom",
		Addr:        "127.0.0.1:10002",
		C:           make(chan string, 100),
		JoinedRooms: make(map[int64]struct{}),
	}

	// 同一个用户的两个 session 上线
	s.Online(tcpUser)
	s.Online(wsUser)

	// 创建运行时房间
	r := room.NewRoom(
		roomID,
		"golang",
	)

	s.AddRoom(r)

	// 这里只让 TCP session 加入房间
	s.JoinRoom(
		tcpUser,
		roomID,
	)

	// 通过 TCP session 发起改名
	s.Rename(
		tcpUser,
		"Jerry",
	)

	// 1. 昵称唯一权威来源 Profiles 应该更新
	profile, ok := s.Profiles[1]
	if !ok {
		t.Fatal("expected user profile")
	}

	if profile.Nickname != "Jerry" {
		t.Fatalf(
			"expected profile nickname Jerry, got %s",
			profile.Nickname,
		)
	}

	// 2. 两个在线 session 都应该还存在
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

	// 3. 改名不应该影响房间成员关系
	gotRoom, ok := s.Rooms[roomID]
	if !ok {
		t.Fatal("expected room exists")
	}

	roomUsers, ok := gotRoom.Users[1]
	if !ok {
		t.Fatal("expected user id 1 in room")
	}

	// 因为只有 tcpUser 加入了，所以这里应该只有一个 session
	if len(roomUsers) != 1 {
		t.Fatalf(
			"expected 1 session in room, got %d",
			len(roomUsers),
		)
	}

	if roomUsers[0] != tcpUser {
		t.Fatal("expected tcp session in room")
	}
}

