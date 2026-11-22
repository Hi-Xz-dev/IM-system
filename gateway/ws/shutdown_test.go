package ws

import (
	"bytes"
	"errors"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"IM-system/internal/connection"
	"IM-system/internal/logger"
	"IM-system/server"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestWebSocketRejectsOversizedAuthMessage(t *testing.T) {
	previous := logger.Log
	logger.Init()
	t.Cleanup(func() { logger.Log = previous })
	s := server.NewServer("", 0, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/ws", NewGateway(s, nil).Handler)
	httpServer := httptest.NewServer(router)
	defer httpServer.Close()
	defer s.Shutdown()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, bytes.Repeat([]byte("a"), connection.MaxMessageBytes+1)); err != nil {
		t.Fatal(err)
	}
	_, _, err = conn.ReadMessage()
	if !websocket.IsCloseError(err, websocket.CloseMessageTooBig) {
		t.Fatalf("expected message-too-big close frame, got %v", err)
	}
}

func TestWebSocketIdleAuthTimesOut(t *testing.T) {
	previous := logger.Log
	logger.Init()
	t.Cleanup(func() { logger.Log = previous })
	s := server.NewServer("", 0, nil, nil, nil, nil)
	g := NewGateway(s, nil)
	router := gin.New()
	router.GET("/ws", g.Handler)
	httpServer := httptest.NewServer(router)
	defer httpServer.Close()
	defer s.Shutdown()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(connection.AuthReadTimeout + 2*time.Second))
	_, payload, err := conn.ReadMessage()
	if err != nil || !strings.Contains(string(payload), "认证失败") {
		t.Fatalf("expected authentication timeout response, got %q, %v", payload, err)
	}
	_, _, err = conn.ReadMessage()
	var netErr net.Error
	if err == nil || (errors.As(err, &netErr) && netErr.Timeout()) {
		t.Fatalf("expected server to disconnect after authentication timeout, got %v", err)
	}
}

func TestShutdownClosesWebSocketWaitingForAuth(t *testing.T) {
	previous := logger.Log
	logger.Init()
	t.Cleanup(func() { logger.Log = previous })
	s := server.NewServer("", 0, nil, nil, nil, nil)
	g := NewGateway(s, nil)
	router := gin.New()
	handlerDone := make(chan struct{})
	router.GET("/ws", func(c *gin.Context) {
		defer close(handlerDone)
		g.Handler(c)
	})
	httpServer := httptest.NewServer(router)
	defer httpServer.Close()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	stopped := make(chan struct{})
	go func() {
		s.Shutdown()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("server shutdown stalled waiting for the WebSocket")
	}

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, _, err = conn.ReadMessage()
	var netErr net.Error
	if err == nil || (errors.As(err, &netErr) && netErr.Timeout()) {
		t.Fatalf("expected WebSocket to be disconnected, got %v", err)
	}
	select {
	case <-handlerDone:
	case <-time.After(time.Second):
		t.Fatal("WebSocket handler did not exit")
	}
}
