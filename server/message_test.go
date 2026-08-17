package server

import (
	"IM-system/internal/domain"
	"IM-system/internal/logger"
	"IM-system/internal/protocol"
	"IM-system/user"

	"testing"
	"time"
)

type fakeConnection struct {
	messages chan string
}

func (f *fakeConnection) Write(data []byte) error {
	f.messages <- string(data)
	return nil
}

func (f *fakeConnection) Close() error {
	return nil
}

func TestPrivateChat(t *testing.T) {

	logger.Init()

	aliceConn := &fakeConnection{
		messages: make(chan string, 10),
	}

	bobConn := &fakeConnection{
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
	
	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
	)

	s.mapLock.Lock()

	s.OnlineUsers[alice.ID] = []*user.User{
		alice,
	}

	s.OnlineUsers[bob.ID] = []*user.User{
		bob,
	}

	s.Profiles[alice.ID] = &UserProfile{
		ID:       alice.ID,
		Nickname: alice.Nickname,
	}

	s.Profiles[bob.ID] = &UserProfile{
		ID:       bob.ID,
		Nickname: bob.Nickname,
	}

	s.mapLock.Unlock()

	s.Online(alice)

	s.Online(bob)

	disconnect := make(chan *user.User, 1)

	go bob.ListenMessage(disconnect)

	s.PrivateChat(
		alice,
		2,
		"hello",
	)

	select {

	case data := <-bobConn.messages:

		msg, err := protocol.DecodeMessage(
			[]byte(data),
		)

		if err != nil {
			t.Fatalf(
				"decode message failed: %v",
				err,
			)
		}

		if msg.Type != domain.MessagePrivate {
			t.Fatalf(
				"expected private message, got %v",
				msg.Type,
			)
		}

		if msg.From != 1 {
			t.Fatalf(
				"expected from 1, got %d",
				msg.From,
			)
		}

		if msg.FromNickname != "Alice" {
			t.Fatalf(
				"expected Alice, got %s",
				msg.FromNickname,
			)
		}

		if msg.To != 2 {
			t.Fatalf(
				"expected to 2, got %d",
				msg.To,
			)
		}

		if msg.Content != "hello" {
			t.Fatalf(
				"expected hello, got %s",
				msg.Content,
			)
		}

	case <-time.After(time.Second):

		t.Fatal("message timeout")
	}
}
