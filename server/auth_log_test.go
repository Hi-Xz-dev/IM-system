package server

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"IM-system/internal/auth"
	"IM-system/internal/connection"
	"IM-system/internal/logger"
)

func TestTCPPartialAuthTimesOut(t *testing.T) {
	var logs bytes.Buffer
	previous := logger.Log
	logger.Log = slog.New(slog.NewJSONHandler(&logs, nil))
	t.Cleanup(func() { logger.Log = previous })
	s := NewServer("", 0, auth.NewService(nil, auth.NewJWTService("test-key")), nil, nil, nil)
	accepted, client := net.Pipe()
	defer client.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.Handler(accepted)
	}()
	defer func() {
		_ = client.Close()
		<-done
	}()

	_ = client.SetDeadline(time.Now().Add(connection.AuthReadTimeout + 2*time.Second))
	if _, err := io.WriteString(client, "auth|partial"); err != nil {
		t.Fatal(err)
	}
	response, err := io.ReadAll(client)
	if err != nil || !strings.Contains(string(response), "认证失败") {
		t.Fatalf("expected authentication failure and server closing the connection, got %q, %v", response, err)
	}
	<-done
	if !strings.Contains(logs.String(), `"reason":"read_failed"`) {
		t.Fatalf("partial input reached authentication instead of failing at read: %s", logs.String())
	}
}

func TestTCPAuthLogOmitsCredentials(t *testing.T) {
	var logs bytes.Buffer
	previous := logger.Log
	logger.Log = slog.New(slog.NewJSONHandler(&logs, nil))
	t.Cleanup(func() { logger.Log = previous })
	s := NewServer("", 0, auth.NewService(nil, auth.NewJWTService("test-key")), nil, nil, testMessagePublisher{})
	for _, input := range []string{
		"auth|sensitive-token-marker",
		"auth|sensitive-token-marker|extra",
		"sensitive-token-marker",
	} {
		logs.Reset()
		if _, _, err := s.authenticateConnection(connection.NewTCPReader(strings.NewReader(input + "\n"))); err == nil {
			t.Fatal("expected authentication failure")
		}
		output := logs.String()
		if strings.Contains(output, "sensitive-token-marker") || strings.Contains(output, `"raw"`) || strings.Contains(output, `"args"`) {
			t.Fatalf("credentials leaked into log: %s", output)
		}
		if !strings.Contains(output, "authentication failed") || !strings.Contains(output, `"protocol":"tcp"`) {
			t.Fatalf("missing safe authentication result: %s", output)
		}
	}
}
