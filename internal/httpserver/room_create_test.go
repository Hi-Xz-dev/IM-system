package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"IM-system/internal/logger"
	"IM-system/internal/model"
	"IM-system/internal/repository"
	"IM-system/server"

	"github.com/gin-gonic/gin"
)

type roomCreateRepository struct {
	created bool
	failure error
}

func (r *roomCreateRepository) Create(_ context.Context, room *model.Room) error {
	if r.failure != nil {
		return r.failure
	}
	room.ID = 1
	r.created = true
	return nil
}

func (r *roomCreateRepository) FindAll(context.Context) ([]model.Room, error) { return nil, nil }

func TestHTTPCreateRoomWithoutChatSession(t *testing.T) {
	logger.Init()
	for _, tc := range []struct {
		name    string
		failure error
		want    int
	}{
		{name: "create empty room", want: http.StatusCreated},
		{name: "duplicate name", failure: repository.ErrRoomAlreadyExists, want: http.StatusConflict},
		{name: "storage failure", failure: errors.New("unavailable"), want: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 注入已认证身份，只验证创建接口；不模拟 JWT 或聊天连接。
			repo := &roomCreateRepository{failure: tc.failure}
			s := server.NewServer("", 0, nil, repo, nil)
			h := NewHandler(s, nil, nil, nil)
			router := gin.New()
			router.POST("/api/rooms", func(c *gin.Context) { c.Set("user_id", int64(1)); h.CreateRoom(c) })
			req := httptest.NewRequest(http.MethodPost, "/api/rooms", strings.NewReader(`{"name":"Go"}`))
			req.Header.Set("Content-Type", "application/json")
			result := httptest.NewRecorder()
			router.ServeHTTP(result, req)
			if result.Code != tc.want {
				t.Fatalf("got %d: %s", result.Code, result.Body.String())
			}
			if tc.failure == nil {
				rooms := s.GetRooms()
				if !repo.created || len(rooms) != 1 || rooms[0].Count != 0 {
					t.Fatal("must publish a persistent empty room")
				}
			} else if len(s.GetRooms()) != 0 || repo.created {
				t.Fatal("must not publish a room after insert failure")
			}
		})
	}
}
