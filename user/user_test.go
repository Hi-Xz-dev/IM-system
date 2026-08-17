package user

import (
	"testing"
	"time"
	"errors"
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

func TestUserSendMessage(t *testing.T) {

	conn := &fakeConnection{
		messages: make(chan string, 1),
	}

	u := &User{
		Nickname:    "Tom",
		ID:          1,
		C:           make(chan string, 100),
		conn:        conn,
		JoinedRooms: make(map[int64]struct{}),
	}

	disconnect := make(chan *User, 1)

	go u.ListenMessage(disconnect)

	err := u.SendMsg("hello")

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	select {

	case msg := <-conn.messages:

		if msg != "hello\n" {
			t.Fatalf(
				"expected hello, got %s",
				msg,
			)
		}

	case <-time.After(time.Second):

		t.Fatal(
			"message timeout",
		)
	}

	close(u.C)
}

func TestUserSendMessageQueueFull(t *testing.T) {

	u := &User{
		Nickname: "Tom",
		C:        make(chan string, 1),
	}

	// 先塞满
	u.C <- "first"

	err := u.SendMsg("second")

	if err == nil {
		t.Fatal(
			"expected queue full error",
		)
	}
}

type errorConnection struct{}

func (e *errorConnection) Write(data []byte) error {
	return errors.New("connection closed")
}

func (e *errorConnection) Close() error {
	return nil
}
func TestListenMessage(t *testing.T) {

	conn := &errorConnection{}

	u := NewUser(
		conn,
		1,
		"Tom",
		"addr",
	)

	disconnect := make(chan *User, 1)

	go u.ListenMessage(disconnect)

	u.SendMsg("hello")

	select {

	case got := <-disconnect:

		if got != u {
			t.Fatal("wrong user")
		}

	case <-time.After(time.Second):

		t.Fatal("disconnect timeout")
	}
}
