package server

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"IM-system/internal/domain"
	"IM-system/internal/logger"
	"IM-system/internal/message"
	"IM-system/internal/protocol"
	"IM-system/internal/repository"
	"IM-system/room"
)

func TestHistoryCommandErrorsHaveStableCodes(t *testing.T) {
	logger.Init()
	db, err := sql.Open("mysql", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	messages := message.NewMessageService(repository.NewMessageRepository(db))
	s := NewServer("", 0, nil, nil, messages, nil)
	s.AddRoom(room.NewRoom(5, "Go"))
	usr := newRoomTestSession(1)

	for _, tc := range []struct {
		input string
		cmd   domain.CommandType
		code  string
	}{
		{"history-room", domain.CmdRoomHistory, "invalid_arguments"},
		{"history-private|0", domain.CmdPrivateHistory, "invalid_arguments"},
		{"history-room|4", domain.CmdRoomHistory, "room_not_found"},
		{"history-room|5", domain.CmdRoomHistory, "not_room_member"},
		{"history-private|2", domain.CmdPrivateHistory, "history_query_failed"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			s.DoMessage(usr, tc.input)
			if len(usr.C) != 1 {
				t.Fatalf("expected one error response, got %d", len(usr.C))
			}
			var response protocol.ErrorResponse
			if err := json.Unmarshal([]byte(<-usr.C), &response); err != nil {
				t.Fatal(err)
			}
			if response.Type != "error" || response.Command != tc.cmd || response.Code != tc.code || response.Message == "" {
				t.Fatalf("unexpected history error response: %+v", response)
			}
		})
	}
}

func TestHistoryResponsePreservesOriginalMessages(t *testing.T) {
	s := NewServer("", 0, nil, nil, nil, nil)
	usr := newRoomTestSession(1)
	older := domain.NewMessage(domain.MessageRoom, 2, "Tom", "older")
	older.Time = older.Time.UTC().Truncate(time.Second)
	older.RoomID = 5
	newer := domain.NewMessage(domain.MessageRoom, 2, "Tom", "newer")
	newer.Time = older.Time.Add(time.Second)
	newer.RoomID = 5
	s.sendHistory(usr, protocol.HistoryResponse{
		RoomID: 5, Messages: []domain.Message{newer, older},
	})
	if len(usr.C) != 1 {
		t.Fatalf("expected one history response, got %d", len(usr.C))
	}

	var got protocol.HistoryResponse
	if err := json.Unmarshal([]byte(<-usr.C), &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "history" || got.RoomID != 5 || len(got.Messages) != 2 {
		t.Fatalf("unexpected history response: %+v", got)
	}
	if got.Messages[0] != older || got.Messages[1] != newer {
		t.Fatalf("history lost original fields or chronological order: %+v", got.Messages)
	}
}

func TestEmptyPrivateHistoryReturnsAnEmptyArray(t *testing.T) {
	s := NewServer("", 0, nil, nil, nil, nil)
	usr := newRoomTestSession(1)
	s.sendHistory(usr, protocol.HistoryResponse{PeerID: 2})
	if len(usr.C) != 1 {
		t.Fatalf("expected one empty history response, got %d", len(usr.C))
	}
	var got protocol.HistoryResponse
	if err := json.Unmarshal([]byte(<-usr.C), &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "history" || got.PeerID != 2 || got.Messages == nil || len(got.Messages) != 0 {
		t.Fatalf("unexpected empty history response: %+v", got)
	}
}
