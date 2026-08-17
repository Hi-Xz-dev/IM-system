package server

import (
	"testing"

	"IM-system/internal/logger"
	"IM-system/room"
	"IM-system/user"
)

// 用户加入房间时，状态维护成本是多少
func BenchmarkJoinRoomUnsafe(b *testing.B) {
	logger.Init()

	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1
	const roomName = "room1"
	const roomSize = 1000

	r := room.NewRoom(
		roomID,
		roomName,
	)

	s.AddRoom(r)

	for i := range roomSize {
		u := &user.User{
			ID:          int64(i + 1),
			Nickname:    "user",
			JoinedRooms: make(map[int64]struct{}),
		}

		s.mapLock.Lock()
		s.joinRoomUnsafe(u, roomID)
		s.mapLock.Unlock()
	}

	id := int64(roomSize + 1)

	for b.Loop() {
		u := &user.User{
			ID:          id,
			Nickname:    "benchmark-user",
			JoinedRooms: make(map[int64]struct{}),
		}

		id++

		s.mapLock.Lock()
		s.joinRoomUnsafe(u, roomID)
		s.mapLock.Unlock()
	}
}

func BenchmarkLeaveJoinRoomUnsafe(b *testing.B) {
	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1

	r := room.NewRoom(
		roomID,
		"room1",
	)
	s.AddRoom(r)

	creator := &user.User{
		ID:          1,
		Nickname:    "creator",
		JoinedRooms: make(map[int64]struct{}),
	}

	u := &user.User{
		ID:          2,
		Nickname:    "benchmark-user",
		JoinedRooms: make(map[int64]struct{}),
	}

	s.mapLock.Lock()
	s.joinRoomUnsafe(u, roomID)
	s.joinRoomUnsafe(creator, roomID)
	s.mapLock.Unlock()

	for b.Loop() {
		s.mapLock.Lock()

		s.leaveRoomUnsafe(u, roomID)
		s.joinRoomUnsafe(u, roomID)

		s.mapLock.Unlock()
	}
}
