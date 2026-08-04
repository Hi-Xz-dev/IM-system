package server

import (
	"testing"
	"time"

	"IM-system/room"
	"IM-system/user"
)

func TestRoomJoinLeave(t *testing.T) {
	s := NewServer("127.0.0.1", 8080, nil)

	u := &user.User{
		ID:          1,
		Nickname:    "Tom",
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

	users, ok := r.Users[u.ID]

	if !ok {
		t.Fatalf("expected Tom in room")
	}

	if len(users) != 1 {
		t.Fatalf("expected 1 user in room, got %d", len(users))
	}

	if users[0] != u {
		t.Fatalf(
			"expected same user",
		)
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

func TestRoomChat(t *testing.T) {
	aliceConn := &fakeConnection{
		messages: make(chan string, 10),
	}

	bobConn := &fakeConnection{
		messages: make(chan string, 10),
	}

	tomConn := &fakeConnection{
		messages: make(chan string, 10),
	}
	alice := user.NewUser(
		aliceConn,
		1,
		"Alice",
		"127.0.0.1",
	)

	bob := user.NewUser(
		bobConn,
		2,
		"Bob",
		"127.0.0.2",
	)

	tom := user.NewUser(
		tomConn,
		3,
		"Tom",
		"127.0.0.3",
	)

	alice.JoinedRooms["room1"] = struct{}{}

	bob.JoinedRooms["room1"] = struct{}{}

	r := &room.Room{
		Name:  "room1",
		Users: make(map[int64][]*user.User),
	}

	r.Users[alice.ID] = []*user.User{alice,}

	r.Users[bob.ID] = []*user.User{bob,}

	s := &Server{
		Rooms: make(map[string]*room.Room),
	}

	s.Rooms["room1"] = r

	disconnect := make(chan *user.User, 3)

	go bob.ListenMessage(disconnect)

	go tom.ListenMessage(disconnect)

	s.RoomChat(
		alice,
		"room1",
		"hello",
	)

	select {

	case msg := <-bobConn.messages:

		t.Log(msg)

	case <-time.After(time.Second):

		t.Fatal("bob timeout")
	}

	select {

	case msg := <-tomConn.messages:

		t.Fatalf(
			"tom should not receive: %s",
			msg,
		)

	case <-time.After(200 * time.Millisecond):

	}

}
