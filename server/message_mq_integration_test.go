package server

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"testing"
	"time"

	"IM-system/internal/config"
	"IM-system/internal/database"
	"IM-system/internal/domain"
	"IM-system/internal/message"
	"IM-system/internal/mq"
	"IM-system/internal/protocol"
	"IM-system/internal/repository"
	"IM-system/user"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestChatMQPersistence(t *testing.T) {
	url := os.Getenv("RABBITMQ_TEST_URL")
	if url == "" {
		t.Skip("set RABBITMQ_TEST_URL to run real broker and MySQL integration")
	}
	db, err := database.NewMySQL(config.Load().MySQL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 使用专属队列，不消费或删除业务队列里的消息。
	queue := "im.message.test." + rand.Text()
	publisher, err := mq.NewRabbitMQ(url, queue)
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Close()
	admin, err := amqp.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	ch, err := admin.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()
	defer func() {
		if _, err := ch.QueueDelete(queue, false, false, false); err != nil {
			t.Error(err)
		}
	}()

	consumer, err := mq.NewConsumer(url, queue)
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()
	service := message.NewMessageService(repository.NewMessageRepository(db))
	s := NewServer("", 0, nil, nil, service, publisher)
	sender, target := newRoomTestSession(1), newRoomTestSession(2)
	s.Online(sender)
	if err := s.SendPrivateMessage(sender, []*user.User{target}, target.ID, "mq integration"); err != nil {
		t.Fatal(err)
	}
	delivered, err := protocol.DecodeMessage([]byte(<-target.C))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DELETE FROM messages WHERE message_id = ?", delivered.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 模拟同一条消息重复进入 MQ。
	if err := publisher.Publish(ctx, delivered); err != nil {
		t.Fatal(err)
	}
	processed := make(chan struct{}, 2)
	done := make(chan error, 1)
	go func() {
		done <- consumer.Run(ctx, func(ctx context.Context, msg domain.Message) error {
			if err := service.Persist(ctx, msg); err != nil {
				return err
			}
			processed <- struct{}{}
			return nil
		})
	}()
	// 无论断言是否失败，都等消费者停止后才关闭数据库或删除测试队列。
	defer func() {
		cancel()
		if err := <-done; err != nil && !errors.Is(err, context.Canceled) {
			t.Error(err)
		}
		q, err := ch.QueueInspect(queue)
		if err != nil {
			t.Error(err)
		} else if q.Messages != 0 {
			t.Errorf("queue still contains %d messages after consumer stopped", q.Messages)
		}
	}()
	for range 2 {
		select {
		case <-processed:
		case <-ctx.Done():
			t.Fatal("timed out waiting for two persist operations")
		}
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages WHERE message_id = ?", delivered.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("duplicate delivery created %d database rows", count)
	}
}
