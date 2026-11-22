package ws

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHeartbeatKeepsQuietConnectionUntilPongsStop(t *testing.T) {
	const timeout = 500 * time.Millisecond
	result := make(chan error, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			result <- err
			return
		}
		stop, err := startHeartbeat(conn, 50*time.Millisecond, timeout)
		if err == nil {
			_, _, err = conn.ReadMessage()
			_ = conn.Close()
			stop()
		} else {
			_ = conn.Close()
		}
		result <- err
	}))
	defer srv.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	until := time.Now().Add(2 * timeout)
	pingHandler := conn.PingHandler()
	sawPingAfterInitialDeadline := false
	conn.SetPingHandler(func(payload string) error {
		if time.Now().Before(until) {
			return pingHandler(payload)
		}
		sawPingAfterInitialDeadline = true
		return nil
	})
	_, _, _ = conn.ReadMessage()
	if !sawPingAfterInitialDeadline {
		t.Error("quiet connection did not survive past its initial deadline")
	}
	select {
	case err := <-result:
		var netErr net.Error
		if !errors.As(err, &netErr) || !netErr.Timeout() {
			t.Fatalf("expected server read timeout after pongs stopped, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not close the connection after pongs stopped")
	}
}
