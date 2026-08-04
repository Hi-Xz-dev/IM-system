package protocol

import (
	"testing"

	"IM-system/internal/domain"
)

func TestEncodeDecodeMessage(t *testing.T) {

	origin := domain.Message{
		Type: domain.MessagePrivate,
		From: 1,
		To: 2,
		Content: "hello",
	}


	data, err := EncodeMessage(origin)

	if err != nil {
		t.Fatal(err)
	}


	got, err := DecodeMessage(data)

	if err != nil {
		t.Fatal(err)
	}


	if got.Type != origin.Type ||
		got.From != origin.From ||
		got.To != origin.To ||
		got.Content != origin.Content {

		t.Fatalf(
			"message mismatch: %+v != %+v",
			got,
			origin,
		)
	}
}
