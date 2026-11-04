package ws

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"IM-system/internal/auth"
	"IM-system/internal/logger"
	"github.com/gorilla/websocket"
)

func TestWSAuthLogOmitsCredentials(t *testing.T) {
	var logs bytes.Buffer
	previous := logger.Log
	logger.Log = slog.New(slog.NewJSONHandler(&logs, nil))
	t.Cleanup(func() { logger.Log = previous })
	g := NewGateway(nil, auth.NewService(nil, auth.NewJWTService("test-key")))
	for _, input := range []string{
		"auth|sensitive-token-marker",
		"auth|sensitive-token-marker|extra",
		"sensitive-token-marker",
	} {
		logs.Reset()
		finished := make(chan error, 1)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				finished <- err
				return
			}
			defer conn.Close()
			_, _, err = g.Auth(conn)
			finished <- err
		}))
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
		if err != nil {
			srv.Close()
			t.Fatal(err)
		}
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
		err = conn.WriteMessage(websocket.TextMessage, []byte(input))
		if err != nil {
			conn.Close()
			srv.Close()
			t.Fatal(err)
		}
		select {
		case err = <-finished:
			if err == nil {
				t.Error("expected authentication failure")
			}
		case <-time.After(3 * time.Second):
			t.Error("authentication timed out")
		}
		conn.Close()
		srv.Close()
		output := logs.String()
		if strings.Contains(output, "sensitive-token-marker") || strings.Contains(output, `"raw"`) || strings.Contains(output, `"args"`) {
			t.Fatalf("credentials leaked into log: %s", output)
		}
		if !strings.Contains(output, "authentication failed") || !strings.Contains(output, `"protocol":"websocket"`) {
			t.Fatalf("missing safe authentication result: %s", output)
		}
	}
}
