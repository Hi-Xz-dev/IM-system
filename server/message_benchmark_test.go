package server

import (
	"context"
	"testing"

	"IM-system/internal/logger"
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

func createBenchmarkUsers(ctx context.Context, n int) []*user.User {

	users := make([]*user.User, 0, n)

	for i := 0; i < n; i++ {

		u := &user.User{
			ID:          int64(i),
			Nickname:    "user",
			C:           make(chan string, 1000),
			JoinedRooms: make(map[string]struct{}),
		}

		startConsumer(
			ctx,
			u,
		)

		users = append(users, u)
	}
	return users
}
//给1000人广播一次消息需要多少成本？
func BenchmarkBroadcastMessage(b *testing.B) {

	logger.Init()

	ctx := b.Context()

	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
	)

	users := createBenchmarkUsers(ctx, 1000)

	sender := users[0]

	s.CreateRoom(
		sender,
		"room1",
	)

	for i := 1; i < len(users); i++ {

		s.JoinRoom(
			users[i],
			"room1",
		)

	}

	s.mapLock.RLock()

	roomUsers := s.getRoomUsersUnsafe("room1")

	s.mapLock.RUnlock()


	for b.Loop() {

		err := s.SendRoomMessage(
			sender,
			roomUsers,
			"room1",
			"hello",
		)

		if err != nil {
			b.Fatal(err)
		}

	}
}
