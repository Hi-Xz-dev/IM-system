package server

import (
	"context"
	"testing"

	"IM-system/internal/logger"
	"IM-system/room"
	"IM-system/user"
)

func startConsumer(ctx context.Context, u *user.User) {

	go func() {

		for {

			select {

			case msg := <-u.C:
				_ = msg

			case <-ctx.Done():
				return
			}
		}
	}()
}

func createBenchmarkUsers(
	ctx context.Context,
	s *Server,
	n int,
) []*user.User {

	users := make([]*user.User, 0, n)

	for i := range n {
		u := &user.User{
			ID:          int64(i),
			Nickname:    "user",
			C:           make(chan string, 1000),
			JoinedRooms: make(map[int64]struct{}),
		}

		startConsumer(
			ctx,
			u,
		)

		s.Profiles[u.ID] = &UserProfile{
			ID:       u.ID,
			Nickname: u.Nickname,
		}

		users = append(users, u)
	}

	return users
}

// 给1000人广播一次消息需要多少成本？
func BenchmarkBroadcastMessage(b *testing.B) {
	logger.Init()

	ctx := b.Context()

	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	const roomID int64 = 1
	const roomName = "room1"

	users := createBenchmarkUsers(
		ctx,
		s,
		1000,
	)

	sender := users[0]

	// 只做 benchmark setup，不走数据库
	r := room.NewRoom(
		roomID,
		roomName,
	)

	s.AddRoom(r)

	// sender 先加入
	s.JoinRoom(
		sender,
		roomID,
	)

	// 其余用户加入
	for i := 1; i < len(users); i++ {
		s.JoinRoom(
			users[i],
			roomID,
		)
	}

	s.mapLock.RLock()

	roomUsers := s.getRoomUsersUnsafe(
		roomID,
	)

	s.mapLock.RUnlock()

	for b.Loop() {
		err := s.SendRoomMessage(
			sender,
			roomUsers,
			roomID,
			roomName,
			"hello",
		)

		if err != nil {
			b.Fatal(err)
		}
	}
}
