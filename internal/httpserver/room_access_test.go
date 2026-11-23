package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"IM-system/internal/auth"
	"IM-system/internal/domain"
	"IM-system/internal/middleware"
	"IM-system/room"
	"IM-system/server"
	"IM-system/user"
	"github.com/gin-gonic/gin"
)

type roomAccessPublisher struct {
	published []domain.Message
}

func (p *roomAccessPublisher) Publish(_ context.Context, msg domain.Message) error {
	p.published = append(p.published, msg)
	return nil
}

func TestRoomAccessFollowsAccountMembership(t *testing.T) {
	publisher := &roomAccessPublisher{}
	s := server.NewServer("", 0, nil, nil, nil, publisher)
	s.AddRoom(room.NewRoom(1, "Go"))
	router := gin.New()
	jwt := auth.NewJWTService("room-access-test-key")
	RegisterRoutes(router, s, nil, nil, middleware.NewAuthMiddleware(jwt), nil, nil)
	var token string
	request := func(method, path string, want int) string {
		t.Helper()
		result := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/api"+path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		router.ServeHTTP(result, req)
		if result.Code != want {
			t.Fatalf("%s returned %d, want %d: %s", path, result.Code, want, result.Body.String())
		}
		return result.Body.String()
	}
	check := func(path string, want int) string {
		t.Helper()
		return request(http.MethodGet, path, want)
	}
	checkDenied := func() {
		check("/rooms/1/members", http.StatusForbidden)
		check("/rooms/1/history", http.StatusForbidden)
	}
	newSession := func() *user.User { return user.NewUser(nil, 7, "Jerry", "") }
	first, second := newSession(), newSession()
	s.Online(first)
	s.Online(second)

	request(http.MethodGet, "/me/rooms", http.StatusUnauthorized)
	request(http.MethodPut, "/me/nickname", http.StatusUnauthorized)
	request(http.MethodPost, "/rooms/1/membership", http.StatusUnauthorized)
	request(http.MethodDelete, "/rooms/1/membership", http.StatusUnauthorized)
	var err error
	token, err = jwt.GenerateToken(7)
	if err != nil {
		t.Fatal(err)
	}
	request(http.MethodPut, "/me/nickname", http.StatusBadRequest)

	checkDenied()
	check("/rooms/1/history?before=invalid!", http.StatusBadRequest)
	check("/rooms/99/history", http.StatusNotFound)
	request(http.MethodPost, "/rooms/1/membership", http.StatusOK)
	request(http.MethodPost, "/rooms/1/members/999", http.StatusOK)
	if body := check("/me/rooms", http.StatusOK); !strings.Contains(body, `"data":["Go"]`) {
		t.Fatalf("expected the authenticated user's room: %s", body)
	}
	s.DoMessage(second, "join|1")
	if len(publisher.published) != 1 || publisher.published[0].From != 7 ||
		publisher.published[0].RoomID != 1 || publisher.published[0].Content != "加入房间" {
		t.Fatalf("expected one join notification from the authenticated account: %+v", publisher.published)
	}
	body := check("/rooms/1/members", http.StatusOK)
	if strings.Count(body, `"nickname":"Jerry"`) != 1 {
		t.Fatalf("expected one account entry for two sessions: %s", body)
	}
	s.Offline(first)
	check("/rooms/1/members", http.StatusOK)

	request(http.MethodDelete, "/rooms/1/membership", http.StatusOK)
	s.DoMessage(second, "leave|1")
	if len(publisher.published) != 2 || publisher.published[1].From != 7 ||
		publisher.published[1].RoomID != 1 || publisher.published[1].Content != "离开房间" {
		t.Fatalf("expected one leave notification from the authenticated account: %+v", publisher.published)
	}
	checkDenied()
	if body := check("/me/rooms", http.StatusOK); !strings.Contains(body, `"data":[]`) {
		t.Fatalf("expected no rooms after leaving: %s", body)
	}

	if msg, ok := s.JoinRoomH(7, 1); !ok {
		t.Fatal(msg)
	}
	s.Offline(second)
	s.Online(newSession())
	checkDenied()
}
