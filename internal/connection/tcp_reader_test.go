package connection

import (
	"strings"
	"testing"
)

func TestTCPReaderMessageLimit(t *testing.T) {
	payload := strings.Repeat("a", MaxMessageBytes)
	reader := NewTCPReader(strings.NewReader(payload + "\r\n" + payload + "b\n"))
	if got, err := reader.Read(); err != nil || got != payload {
		t.Fatalf("limit-sized CRLF message: read %d bytes, err=%v", len(got), err)
	}
	if _, err := reader.Read(); err == nil {
		t.Fatal("accepted a message exceeding the limit by one byte")
	}
}
