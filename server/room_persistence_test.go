package server

import (
	"context"
	"errors"
	"testing"

	"IM-system/internal/model"
	"IM-system/internal/repository"
)

type memoryRoomRepository struct {
	rooms        []model.Room
	failure      error
	beforeCreate func()
}

func (r *memoryRoomRepository) Create(ctx context.Context, room *model.Room) error {
	if r.failure != nil {
		return r.failure
	}
	if r.beforeCreate != nil {
		r.beforeCreate()
	}
	for _, existing := range r.rooms {
		if existing.Name == room.Name {
			return repository.ErrRoomAlreadyExists
		}
	}
	room.ID = int64(len(r.rooms) + 1)
	r.rooms = append(r.rooms, *room)
	return nil
}

func (r *memoryRoomRepository) FindAll(context.Context) ([]model.Room, error) {
	return append([]model.Room(nil), r.rooms...), r.failure
}

func TestPersistentEmptyRoomsAndRestart(t *testing.T) {
	repo := &memoryRoomRepository{}
	s := NewServer("", 0, nil, repo, nil)
	u := newRoomTestSession(1)
	s.Online(u)
	id, err := s.CreateRoom(u, "Go")
	if err != nil {
		t.Fatal(err)
	}
	s.LeaveRoom(u, id)
	if len(s.Rooms[id].Users) != 0 {
		t.Fatal("leave must preserve an empty room")
	}
	s.JoinRoom(u, id)
	s.Offline(u)
	if len(s.Rooms[id].Users) != 0 {
		t.Fatal("disconnect must preserve an empty room")
	}

	restarted := NewServer("", 0, nil, repo, nil)
	if err := restarted.LoadRooms(context.Background()); err != nil {
		t.Fatal(err)
	}
	rooms := restarted.GetRooms()
	if len(rooms) != 1 || rooms[0].ID != id || rooms[0].Name != "Go" || rooms[0].Count != 0 {
		t.Fatalf("unexpected restored rooms: %+v", rooms)
	}
	reconnected := newRoomTestSession(1)
	restarted.Online(reconnected)
	if reconnected.InRoom(id) {
		t.Fatal("restart must not restore membership")
	}
	if _, err := restarted.CreateRoomForUser(context.Background(), 1, "Go"); !errors.Is(err, repository.ErrRoomAlreadyExists) {
		t.Fatalf("expected existing room conflict, got %v", err)
	}
	restarted.JoinRoom(reconnected, id)
	if !reconnected.InRoom(id) {
		t.Fatal("restored room must be joinable")
	}
}

func TestCreateRoomWithoutChatConnection(t *testing.T) {
	for _, disconnectDuringWrite := range []bool{false, true} {
		t.Run(map[bool]string{false: "offline", true: "disconnect during insert"}[disconnectDuringWrite], func(t *testing.T) {
			repo := &memoryRoomRepository{}
			s := NewServer("", 0, nil, repo, nil)
			if disconnectDuringWrite {
				u := newRoomTestSession(1)
				s.Online(u)
				repo.beforeCreate = func() { s.Offline(u) }
			}
			created, err := s.CreateRoomForUser(context.Background(), 1, "empty")
			if err != nil {
				t.Fatal(err)
			}
			if len(repo.rooms) != 1 || s.Rooms[created.ID] == nil || len(s.Rooms[created.ID].Users) != 0 {
				t.Fatal("successful insert must publish an empty room even without a connection")
			}
		})
	}
}

func TestRoomStorageFailureDoesNotPublishState(t *testing.T) {
	failure := errors.New("storage unavailable")
	s := NewServer("", 0, nil, &memoryRoomRepository{failure: failure}, nil)
	if _, err := s.CreateRoomForUser(context.Background(), 1, "Go"); !errors.Is(err, failure) {
		t.Fatalf("expected insert error, got %v", err)
	}
	if err := s.LoadRooms(context.Background()); !errors.Is(err, failure) {
		t.Fatalf("expected load error, got %v", err)
	}
	if len(s.Rooms) != 0 {
		t.Fatal("storage errors must not publish partial rooms")
	}
}
