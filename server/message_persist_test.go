package server

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"IM-system/internal/domain"
	"IM-system/internal/message"
	"IM-system/internal/repository"
	

	_ "github.com/go-sql-driver/mysql"
)

func TestAsyncPersistMessage(t *testing.T) {
	db, err := sql.Open(
		"mysql",
		"root@tcp(127.0.0.1:3306)/im_system?parseTime=true",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	messageRepo := repository.NewMessageRepository(db)

	messageService := message.NewMessageService(
		messageRepo,
	)

	s := NewServer(
		"127.0.0.1",
		8080,
		nil,
		nil,
		messageService,
	)

	content := "async-persist-test"

	// 先清理上一次可能残留的数据
	if _, err := db.Exec(
		"DELETE FROM messages WHERE content = ?",
		content,
	); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM messages WHERE content = ?",
			content,
		)
		if err != nil {
			t.Error(err)
		}
	})

	// 启动异步持久化 worker
	go s.ListenPersistMessage()

	msg := domain.Message{
		Type:         domain.MessagePrivate,
		From:         1,
		FromNickname: "Alice",
		To:           2,
		Content:      content,
		Time:         time.Now(),
	}

	// 不直接调用 Persist
	// 只把消息放进队列
	s.PersistQueue <- msg

	deadline := time.Now().Add(time.Second)

	for {
		var count int

		err := db.QueryRowContext(
			context.Background(),
			`SELECT COUNT(*)
			 FROM messages
			 WHERE content = ?`,
			content,
		).Scan(&count)

		if err != nil {
			t.Fatal(err)
		}

		if count == 1 {
			break
		}

		if time.Now().After(deadline) {
			t.Fatal("async message persist timeout")
		}

		time.Sleep(10 * time.Millisecond)
	}
}