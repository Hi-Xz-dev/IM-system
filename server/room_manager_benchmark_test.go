package server

import (
	"testing"

	"IM-system/internal/logger"
	"IM-system/user"
)

// 用户加入房间时，状态维护成本是多少
func BenchmarkJoinRoomUnsafe(b *testing.B) {

	logger.Init()

	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
	)

	creator := &user.User{
		ID:          1,
		Nickname:    "creator",
		C:           make(chan string, 1000),
		JoinedRooms: make(map[string]struct{}),
	}

	s.CreateRoom(
		creator,
		"room1",
	)

	const roomSize = 1000

	for i := range roomSize {

		u := &user.User{
			ID:          int64(i + 2),
			Nickname:    "user",
			JoinedRooms: make(map[string]struct{}),
		}

		s.mapLock.Lock()

		s.joinRoomUnsafe(
			u,
			"room1",
		)

		s.mapLock.Unlock()
	}

	id := int64(roomSize + 2)

	for b.Loop() {
		u := &user.User{
			ID:          id,
			Nickname:    "benchmark-user",
			JoinedRooms: make(map[string]struct{}),
		}
		id++

		s.mapLock.Lock()

		s.joinRoomUnsafe(
			u,
			"room1",
		)

		s.mapLock.Unlock()

	}

}

func BenchmarkLeaveRoomUnsafe(b *testing.B) {
	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
	)

	creator := &user.User{
		ID:          1,
		Nickname:    "creator",
		JoinedRooms: make(map[string]struct{}),
	}

	s.CreateRoom(creator, "room1")

	const userCount = 10000

	users := make([]*user.User, userCount)

	for i := range userCount {

		u := &user.User{
			ID:          int64(i + 2),
			Nickname:    "user",
			JoinedRooms: make(map[string]struct{}),
		}

		users[i] = u

		s.mapLock.Lock()
		s.joinRoomUnsafe(u, "room1")
		s.mapLock.Unlock()

	}

	index := 0

	for b.Loop() {

		u := users[index]

		index++

		if index == userCount {
			index = 0
		}
		s.mapLock.Lock()

		s.leaveRoomUnsafe(
			u,
			"room1",
		)

		s.mapLock.Unlock()

	}

}
