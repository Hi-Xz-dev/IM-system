package server

import (
	"context"
	"errors"
	"testing"

	"IM-system/internal/domain"
	"IM-system/internal/protocol"
	"IM-system/user"
)

// 业务测试不依赖本地 broker；真实服务注入 RabbitMQ 发布者。
type testMessagePublisher struct {
	publish func(context.Context, domain.Message) error
}

func (p testMessagePublisher) Publish(ctx context.Context, msg domain.Message) error {
	if p.publish != nil {
		return p.publish(ctx, msg)
	}
	return nil
}

func TestMessagePublishBeforeDelivery(t *testing.T) {
	for _, kind := range []string{"private", "room"} {
		for _, rejected := range []bool{false, true} {
			t.Run(kind+map[bool]string{false: "/accepted", true: "/rejected"}[rejected], func(t *testing.T) {
				sender := newRoomTestSession(1)
				target := newRoomTestSession(2)
				publishErr := errors.New("broker unavailable")
				var published domain.Message
				publisher := testMessagePublisher{publish: func(ctx context.Context, msg domain.Message) error {
					if _, ok := ctx.Deadline(); !ok {
						t.Fatal("publish context has no deadline")
					}
					if len(target.C) != 0 {
						t.Fatal("delivered before publisher confirmation")
					}
					published = msg
					if rejected {
						return publishErr
					}
					return nil
				}}
				s := NewServer("", 0, nil, nil, nil, publisher)
				s.Online(sender)
				var err error
				if kind == "private" {
					err = s.SendPrivateMessage(sender, []*user.User{target}, target.ID, "hello")
				} else {
					err = s.SendRoomMessage(sender, []*user.User{target}, 10, "room", "hello")
				}
				if published.ID == "" || published.Content != "hello" {
					t.Fatalf("unexpected published message: %+v", published)
				}
				if rejected {
					if !errors.Is(err, publishErr) || len(target.C) != 0 {
						t.Fatalf("failed publish must prevent delivery: %v", err)
					}
					return
				}
				if err != nil || len(target.C) != 1 {
					t.Fatalf("confirmed message not delivered: %v", err)
				}
				delivered, err := protocol.DecodeMessage([]byte(<-target.C))
				if err != nil || delivered.ID != published.ID {
					t.Fatalf("published and delivered IDs differ: %v", err)
				}
			})
		}
	}
}

func TestMessagePublishRequiresPublisher(t *testing.T) {
	s := NewServer("", 0, nil, nil, nil, nil)
	if err := s.publishMessage(domain.SystemMessage("hello")); err == nil {
		t.Fatal("missing publisher must fail rather than silently skip persistence")
	}
}
