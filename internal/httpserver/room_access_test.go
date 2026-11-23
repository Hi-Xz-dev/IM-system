package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"IM-system/room"
	"IM-system/server"
	"IM-system/user"
	"github.com/gin-gonic/gin"
)

func TestRoomAccessFollowsAccountMembership(t *testing.T) {
	s := server.NewServer("", 0, nil, nil, nil, nil)
	s.AddRoom(room.NewRoom(1, "Go"))
	h := NewHandler(s, nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", int64(7)) })
	router.GET("/rooms/:room/members", h.Members)
	router.GET("/rooms/:room/history", h.RoomHistory)
	check := func(path string, want int) string {
		t.Helper()
		result := httptest.NewRecorder()
		router.ServeHTTP(result, httptest.NewRequest(http.MethodGet, path, nil))
		if result.Code != want {
			t.Fatalf("%s returned %d, want %d: %s", path, result.Code, want, result.Body.String())
		}
		return result.Body.String()
	}
	checkDenied := func() {
		check("/rooms/1/members", http.StatusForbidden)
		check("/rooms/1/history", http.StatusForbidden)
	}
	newSession := func() *user.User { return user.NewUser(nil, 7, "Jerry", "") }
	first, second := newSession(), newSession()
	s.Online(first)
	s.Online(second)

	checkDenied()
	check("/rooms/99/history", http.StatusNotFound)
	if msg, ok := s.JoinRoomH(7, 1); !ok {
		t.Fatal(msg)
	}
	body := check("/rooms/1/members", http.StatusOK)
	if strings.Count(body, `"nickname":"Jerry"`) != 1 {
		t.Fatalf("expected one account entry for two sessions: %s", body)
	}
	s.Offline(first)
	check("/rooms/1/members", http.StatusOK)

	if msg, ok := s.LeaveRoomH(7, 1); !ok {
		t.Fatal(msg)
	}
	checkDenied()

	if msg, ok := s.JoinRoomH(7, 1); !ok {
		t.Fatal(msg)
	}
	s.Offline(second)
	s.Online(newSession())
	checkDenied()
}
