package protocol

import (
	"testing"

	"IM-system/internal/domain"
)
func TestEncodeMessage(t *testing.T){

	msg := domain.Message{
		Type: domain.MessagePrivate,
		From: 1,
		To:2,
		Content:"hello",
	}


	data,err:=EncodeMessage(msg)

	if err!=nil{
		t.Fatal(err)
	}


	t.Log(string(data))

}