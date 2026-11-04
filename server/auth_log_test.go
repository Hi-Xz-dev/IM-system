package server

import (
	"bufio"
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"IM-system/internal/auth"
	"IM-system/internal/logger"
)

func TestTCPAuthLogOmitsCredentials(t *testing.T) {
	var logs bytes.Buffer
	previous := logger.Log
	logger.Log = slog.New(slog.NewJSONHandler(&logs, nil))
	t.Cleanup(func() { logger.Log = previous })
	s := NewServer("", 0, auth.NewService(nil, auth.NewJWTService("test-key")), nil, nil)
	for _, input := range []string{
		"auth|sensitive-token-marker",
		"auth|sensitive-token-marker|extra",
		"sensitive-token-marker",
	} {
		logs.Reset()
		if _, _, err := s.authenticateConnection(bufio.NewScanner(strings.NewReader(input + "\n"))); err == nil {
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
