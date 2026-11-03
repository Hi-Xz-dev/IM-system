package server

import (
	"testing"

	"IM-system/internal/logger"
	"IM-system/internal/protocol"
	"IM-system/user"
)

func TestPrivateChatDeliveryResult(t *testing.T) {
	logger.Init()
	for _, tc := range []struct {
		name           string
		fullTargets    int
		readyTargets   int
		missingProfile bool
		want           string
	}{
		{name: "offline", want: "用户当前不在线"},
		{name: "all queues full", fullTargets: 2, want: "私聊发送失败，请稍后重试"},
		{name: "partial enqueue", fullTargets: 1, readyTargets: 1, want: "私聊消息已进入发送队列"},
		{name: "all enqueued", readyTargets: 2, want: "私聊消息已进入发送队列"},
		{name: "service error", readyTargets: 1, missingProfile: true, want: "私聊发送失败，请稍后重试"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewServer("127.0.0.1", 0, nil, nil, nil)
			sender := &user.User{ID: 1, C: make(chan string, 10)}
			if !tc.missingProfile {
				s.Profiles[1] = &UserProfile{ID: 1, Nickname: "sender"}
			}
			for range tc.fullTargets {
				u := &user.User{ID: 2, C: make(chan string, 1)}
				u.C <- "occupied"
				s.OnlineUsers[2] = append(s.OnlineUsers[2], u)
			}
			for range tc.readyTargets {
				s.OnlineUsers[2] = append(s.OnlineUsers[2], &user.User{ID: 2, C: make(chan string, 1)})
			}
			s.PrivateChat(sender, 2, "hello")
			if len(sender.C) != 1 {
				t.Fatalf("expected one result notification, got %d", len(sender.C))
			}
			msg, err := protocol.DecodeMessage([]byte(<-sender.C))
			if err != nil {
				t.Fatal(err)
			}
			if msg.Content != tc.want {
				t.Fatalf("got %q, want %q", msg.Content, tc.want)
			}
			for _, target := range s.OnlineUsers[2][tc.fullTargets:] {
				if tc.missingProfile {
					if len(target.C) != 0 {
						t.Fatal("unexpected delivery after service error")
					}
					continue
				}
				if len(target.C) != 1 {
					t.Fatal("ready session did not receive the message")
				}
				delivered, err := protocol.DecodeMessage([]byte(<-target.C))
				if err != nil || delivered.Content != "hello" || delivered.To != 2 {
					t.Fatalf("unexpected delivered message: %+v, error: %v", delivered, err)
				}
			}
		})
	}
}
