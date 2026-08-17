package server

import (
	"sync"
	"testing"
	"time"

	"IM-system/internal/logger"
	"IM-system/room"
	"IM-system/user"
)

func TestRoomJoinLeave(t *testing.T) {

	s := NewServer("127.0.0.1", 8080, nil, nil)

	const roomID int64 = 1

	u := &user.User{
		ID:          1,
		Nickname:    "Tom",
		Addr:        "127.0.0.1:10001",
		C:           make(chan string, 100),
		JoinedRooms: make(map[int64]struct{}),
	}

	s.Online(u)

	r := room.NewRoom(roomID, "golang")

	s.AddRoom(r)

	s.JoinRoom(u, roomID)

	// 创建房间后，用户应该进入该房间
	if _, ok := u.JoinedRooms[roomID]; !ok {
		t.Fatalf("expected user joined room %d", roomID)
	}
	// 房间应该已经创建
	r, ok := s.Rooms[roomID]
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
	s.LeaveRoom(u, roomID)

	// 离开房间后，用户当前房间应该为空
	if _, ok := u.JoinedRooms[roomID]; ok {
		t.Fatalf("expected current room empty, got %d", roomID)
	}

	// 房间为空后，应该自动删除
	if _, ok := s.Rooms[roomID]; ok {
		t.Fatalf("expected room golang deleted")
	}
}

func TestRoomChat(t *testing.T) {
	const roomID int64 = 1

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

	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	// 建立完整运行时状态：
	// OnlineUsers + Profiles
	s.Online(alice)
	s.Online(bob)
	s.Online(tom)

	// 测试 setup：直接创建运行时房间
	r := room.NewRoom(
		roomID,
		"room1",
	)

	s.AddRoom(r)

	// Alice、Bob 加入房间
	s.JoinRoom(alice, roomID)
	s.JoinRoom(bob, roomID)

	disconnect := make(chan *user.User, 3)

	go bob.ListenMessage(disconnect)
	go tom.ListenMessage(disconnect)

	s.RoomChat(
		alice,
		roomID,
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

func TestJoinNonExistentRoom(t *testing.T) {
	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1

	u := &user.User{
		ID:          1,
		Nickname:    "Tom",
		C:           make(chan string, 10),
		JoinedRooms: make(map[int64]struct{}),
	}

	s.Online(u)

	s.JoinRoom(
		u,
		roomID,
	)

	// 用户不应该加入不存在的房间
	if u.InRoom(roomID) {
		t.Fatal("user joined non-existent room")
	}

	// 应该收到系统提示
	select {
	case msg := <-u.C:
		if msg == "" {
			t.Fatal("empty system message")
		}

	case <-time.After(time.Second):
		t.Fatal("expected system message")
	}
}

func TestLeaveNonJoinedRoom(t *testing.T) {
	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1

	u1 := &user.User{
		ID:          1,
		Nickname:    "Tom",
		C:           make(chan string, 10),
		JoinedRooms: make(map[int64]struct{}),
	}

	u2 := &user.User{
		ID:          2,
		Nickname:    "Jerry",
		C:           make(chan string, 10),
		JoinedRooms: make(map[int64]struct{}),
	}

	s.Online(u1)
	s.Online(u2)

	// 只创建运行时房间
	r := room.NewRoom(
		roomID,
		"room1",
	)
	s.AddRoom(r)

	// u1 加入，保证这个房间确实存在且不是空房间
	s.JoinRoom(
		u1,
		roomID,
	)

	// u2 从来没加入过
	s.LeaveRoom(
		u2,
		roomID,
	)

	if u2.InRoom(roomID) {
		t.Fatal("user should not join room")
	}

	r, ok := s.Rooms[roomID]
	if !ok {
		t.Fatal("room should still exist")
	}

	if _, ok := r.Users[u2.ID]; ok {
		t.Fatal("user should not be in room")
	}

	select {
	case msg := <-u2.C:
		if msg == "" {
			t.Fatal("empty system message")
		}

	case <-time.After(time.Second):
		t.Fatal("expected system message")
	}
}

func TestConcurrentRoomJoin(t *testing.T) {
	logger.Init()

	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1
	const n = 10

	creator := &user.User{
		ID:          0,
		Nickname:    "creator",
		C:           make(chan string, 10),
		JoinedRooms: make(map[int64]struct{}),
	}

	s.Online(creator)

	r := room.NewRoom(
		roomID,
		"room1",
	)

	s.AddRoom(r)

	s.JoinRoom(
		creator,
		roomID,
	)

	var wg sync.WaitGroup

	for i := range n {
		wg.Go(func(){

			u := &user.User{
				ID:          int64(i + 1),
				Nickname:    "user",
				C:           make(chan string, 10),
				JoinedRooms: make(map[int64]struct{}),
			}

			s.Online(u)

			s.JoinRoom(
				u,
				roomID,
			)
		})
	}

	wg.Wait()

	s.mapLock.RLock()

	gotRoom, ok := s.Rooms[roomID]
	if !ok {
		s.mapLock.RUnlock()
		t.Fatal("room not found")
	}

	got := len(gotRoom.Users)

	s.mapLock.RUnlock()

	if got != n+1 {
		t.Fatalf(
			"expected %d users, got %d",
			n+1,
			got,
		)
	}
}

func TestConcurrentOffline(t *testing.T) {
	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1

	u := &user.User{
		ID:          1,
		Nickname:    "Tom",
		C:           make(chan string, 100),
		JoinedRooms: make(map[int64]struct{}),
	}

	s.Online(u)

	r := room.NewRoom(
		roomID,
		"room1",
	)

	s.AddRoom(r)

	s.JoinRoom(
		u,
		roomID,
	)

	var wg sync.WaitGroup

	const n = 10

	for range n {
		wg.Go(func() {
			s.Offline(u)
		})
	}

	wg.Wait()

	if !u.IsClosed {
		t.Fatal("user should be closed")
	}

	s.mapLock.RLock()
	defer s.mapLock.RUnlock()

	if _, ok := s.OnlineUsers[u.ID]; ok {
		t.Fatal("user still online")
	}

	if _, ok := s.Rooms[roomID]; ok {
		t.Fatal("room still exists")
	}
}
