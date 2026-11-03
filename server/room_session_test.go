package server

import (
	"sync"
	"testing"

	"IM-system/room"
	"IM-system/user"
)

func newRoomTestSession(id int64) *user.User {
	return &user.User{
		ID: id, Nickname: "test", C: make(chan string, 100),
		JoinedRooms: make(map[int64]struct{}),
	}
}

func setupRoomSessions(t *testing.T) (*Server, *room.Room, *user.User, *user.User, *user.User) {
	t.Helper()
	s := NewServer("127.0.0.1", 0, nil, nil, nil)
	newSession := func(id int64) *user.User {
		u := &user.User{
			ID: id, Nickname: "test", C: make(chan string, 100),
			JoinedRooms: make(map[int64]struct{}),
		}
		s.Online(u)
		return u
	}
	a1, a2, b := newSession(1), newSession(1), newSession(2)
	r := room.NewRoom(1, "test")
	s.AddRoom(r)
	s.mapLock.Lock()
	for _, u := range []*user.User{a1, a2, b} {
		s.joinRoomUnsafe(u, r.ID)
	}
	s.mapLock.Unlock()
	return s, r, a1, a2, b
}

func TestJoinSynchronizesAccountSessions(t *testing.T) {
	for _, entry := range []string{"command", "http", "create"} {
		t.Run(entry, func(t *testing.T) {
			s := NewServer("127.0.0.1", 0, nil, nil, nil)
			a1, a2 := newRoomTestSession(1), newRoomTestSession(1)
			s.Online(a1)
			s.Online(a2)
			if entry == "create" {
				s.roomRepo = &memoryRoomRepository{}
				if _, err := s.CreateRoom(a1, "test"); err != nil {
					t.Fatal(err)
				}
			} else {
				s.AddRoom(room.NewRoom(1, "test"))
				for range 2 { // Repeated joins must not duplicate delivery targets.
					if entry == "command" {
						s.JoinRoom(a1, 1)
					} else if msg, ok := s.JoinRoomH(a1.ID, 1); !ok {
						t.Fatal(msg)
					}
				}
			}
			if !a1.InRoom(1) || !a2.InRoom(1) || len(s.Rooms[1].Users[1]) != 2 {
				t.Fatal("joining must synchronize all online sessions without duplicates")
			}
		})
	}
}

func TestNewSessionInheritsOnlyOnlineMembership(t *testing.T) {
	s, r, a1, a2, b := setupRoomSessions(t)
	a3 := newRoomTestSession(1)
	s.Online(a3)
	if !a3.InRoom(r.ID) || len(r.Users[1]) != 3 {
		t.Fatal("new session must inherit membership once, even with multiple existing sessions")
	}
	s.Offline(a1)
	s.Offline(a2)
	s.Offline(a3)
	if s.Rooms[r.ID] != r || !b.InRoom(r.ID) {
		t.Fatal("another account must keep the room alive")
	}
	a4 := newRoomTestSession(1)
	s.Online(a4)
	if a4.InRoom(r.ID) || len(r.Users[1]) != 0 {
		t.Fatal("fully offline account must not regain old membership")
	}
	s.JoinRoom(a4, r.ID)
	s.LeaveRoom(a4, r.ID)
	a5 := newRoomTestSession(1)
	s.Online(a5)
	if a5.InRoom(r.ID) {
		t.Fatal("a new session must not inherit explicitly abandoned membership")
	}
}

func TestJoinAndNewSessionConcurrent(t *testing.T) {
	for range 10 {
		s := NewServer("127.0.0.1", 0, nil, nil, nil)
		a1, a2 := newRoomTestSession(1), newRoomTestSession(1)
		s.Online(a1)
		s.AddRoom(room.NewRoom(1, "test"))
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			s.JoinRoom(a1, 1)
		}()
		go func() {
			defer wg.Done()
			s.Online(a2)
		}()
		wg.Wait()
		if !a1.InRoom(1) || !a2.InRoom(1) || len(s.Rooms[1].Users[1]) != 2 {
			t.Fatal("join/login ordering must not lose or duplicate membership")
		}
	}
}

func TestExplicitLeaveRemovesAllAccountSessions(t *testing.T) {
	for _, entry := range []string{"command", "http"} {
		t.Run(entry, func(t *testing.T) {
			s, r, a1, a2, b := setupRoomSessions(t)
			if entry == "command" {
				s.LeaveRoom(a2, r.ID)
			} else if msg, ok := s.LeaveRoomH(a2.ID, r.ID); !ok {
				t.Fatal(msg)
			}
			if a1.InRoom(r.ID) || a2.InRoom(r.ID) {
				t.Fatal("explicit leave must clear both sessions' room membership")
			}
			if _, ok := r.Users[a1.ID]; ok {
				t.Fatal("account still present in room")
			}
			if len(s.OnlineUsers[a1.ID]) != 2 || a1.IsClosed || a2.IsClosed {
				t.Fatal("leaving a room must not disconnect account sessions")
			}
			if s.Rooms[r.ID] != r || len(r.Users[b.ID]) != 1 || !b.InRoom(r.ID) {
				t.Fatal("another account's membership was changed")
			}
		})
	}
}

func TestOfflinePreservesOtherRoomSessions(t *testing.T) {
	s, r, a1, a2, b := setupRoomSessions(t)
	s.Offline(a1)
	s.Offline(a1) // Repeated disconnect cleanup must be harmless.
	if !a1.IsClosed || a1.InRoom(r.ID) {
		t.Fatal("disconnected session was not cleaned up")
	}
	if a2.IsClosed || !a2.InRoom(r.ID) {
		t.Fatal("disconnect must preserve the other session's membership")
	}
	if sessions := r.Users[a1.ID]; len(sessions) != 1 || sessions[0] != a2 {
		t.Fatal("room must retain exactly the surviving account session")
	}
	if sessions := s.OnlineUsers[a1.ID]; len(sessions) != 1 || sessions[0] != a2 {
		t.Fatal("online account must retain the surviving session")
	}
	s.RoomChat(b, r.ID, "still connected")
	if len(a2.C) == 0 {
		t.Fatal("surviving session did not receive a room message")
	}
	s.Offline(a2)
	if _, ok := r.Users[a1.ID]; ok {
		t.Fatal("last account session must remove the account from the room")
	}
	if s.Rooms[r.ID] != r || !b.InRoom(r.ID) {
		t.Fatal("room with another account must remain")
	}
	s.Offline(b)
	if s.Rooms[r.ID] != r || len(r.Users) != 0 {
		t.Fatal("empty room must remain after all sessions disconnect")
	}
}

func TestOfflineThenExplicitLeave(t *testing.T) {
	s, r, a1, a2, _ := setupRoomSessions(t)
	s.Offline(a1)
	if msg, ok := s.LeaveRoomH(a2.ID, r.ID); !ok {
		t.Fatal(msg)
	}
	if a2.InRoom(r.ID) {
		t.Fatal("remaining session must be able to explicitly leave")
	}
	if _, ok := r.Users[a2.ID]; ok {
		t.Fatal("account still present after explicit leave")
	}
}
