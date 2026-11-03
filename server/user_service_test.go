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

	// TCP session 发起加入，同账号所有在线端同步加入。
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

	// 改名后两个端都应保留房间成员关系。
	if len(roomUsers) != 2 {
		t.Fatalf(
			"expected 2 sessions in room, got %d",
			len(roomUsers),
		)
	}

	if roomUsers[0] != tcpUser || roomUsers[1] != wsUser {
		t.Fatal("expected both tcp and websocket sessions in room")
	}
	if !tcpUser.InRoom(roomID) || !wsUser.InRoom(roomID) {
		t.Fatal("rename must preserve membership on both sessions")
	}
}

