package message

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"IM-system/internal/domain"
	"IM-system/internal/model"
	"IM-system/internal/repository"

	_ "github.com/go-sql-driver/mysql"
)

func TestHistoryPagesWithSameTimestamp(t *testing.T) {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/im_system?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := NewMessageService(repository.NewMessageRepository(db))
	ctx := context.Background()
	for _, kind := range []domain.MessageType{domain.MessageRoom, domain.MessagePrivate} {
		t.Run(string(kind), func(t *testing.T) {
			// Use IDs separate from the other fixtures and clean up by message ID.
			sender := time.Now().UnixNano()
			recipient := sender + 1
			when := time.Now().UTC().Truncate(time.Second)
			insert := func(i int) string {
				t.Helper()
				msg := domain.NewMessage(kind, sender, "sender", fmt.Sprint(i))
				if kind == domain.MessageRoom {
					msg.RoomID = recipient
				} else {
					msg.To = recipient
				}
				msg.Time = when
				if err := service.Persist(ctx, msg); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if _, err := db.Exec("DELETE FROM messages WHERE message_id = ?", msg.ID); err != nil {
						t.Error(err)
					}
				})
				return msg.ID
			}
			load := func(before *model.HistoryPosition) HistoryPage {
				t.Helper()
				var page HistoryPage
				var err error
				if kind == domain.MessageRoom {
					page, err = service.GetRoomHistory(ctx, recipient, before)
				} else {
					page, err = service.GetPrivateHistory(ctx, sender, recipient, before)
				}
				if err != nil {
					t.Fatal(err)
				}
				return page
			}
			const total = historyPageSize + 2
			ids := make([]string, total)
			for i := range ids {
				ids[i] = insert(i)
			}
			first := load(nil)
			if len(first.Messages) != historyPageSize || first.NextCursor == "" {
				t.Fatalf("expected a full first page and cursor, got %d messages", len(first.Messages))
			}
			before, err := ParseHistoryCursor(first.NextCursor)
			if err != nil {
				t.Fatal(err)
			}
			insert(total) // A new message must not shift the second page.
			second := load(before)
			if len(second.Messages) != 2 || second.NextCursor != "" {
				t.Fatalf("expected two remaining messages and no cursor, got %d messages", len(second.Messages))
			}
			for i, msg := range append(first.Messages, second.Messages...) {
				if msg.ID != ids[total-1-i] {
					t.Fatalf("history message %d had wrong ID or order", i)
				}
			}
		})
	}
}

func TestMessageServicePersistRoom(t *testing.T) {
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
	messageService := NewMessageService(messageRepo)

	content := "test-room-message"

	// 先清理旧测试数据
	if _, err := db.Exec(
		"DELETE FROM messages WHERE content = ?",
		content,
	); err != nil {
		t.Fatal(err)
	}

	msg := domain.Message{
		ID:           rand.Text(),
		Type:         domain.MessageRoom,
		From:         1,
		FromNickname: "Tom",
		RoomID:       1,
		RoomName:     "golang",
		Content:      content,
		Time:         time.Now(),
	}

	// 1. 持久化
	if err := messageService.Persist(
		context.Background(),
		msg,
	); err != nil {
		t.Fatal(err)
	}

	history, err := messageService.GetRoomHistory(context.Background(), msg.RoomID, nil)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range history.Messages {
		if entry.ID == msg.ID && entry.Content == msg.Content {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("history did not preserve live message ID %q", msg.ID)
	}

	// 2. 从数据库查回来
	var (
		id           int64
		msgType      string
		fromID       int64
		fromNickname string

		toID     sql.NullInt64
		roomID   sql.NullInt64
		roomName sql.NullString

		gotContent string
	)

	err = db.QueryRow(
		`SELECT
			id,
			type,
			from_id,
			from_nickname,
			to_id,
			room_id,
			room_name,
			content
		FROM messages
		WHERE content = ?
		ORDER BY id DESC
		LIMIT 1`,
		content,
	).Scan(
		&id,
		&msgType,
		&fromID,
		&fromNickname,
		&toID,
		&roomID,
		&roomName,
		&gotContent,
	)

	if err != nil {
		t.Fatal(err)
	}

	// 3. 测试结束后删除这条消息
	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM messages WHERE id = ?",
			id,
		)
		if err != nil {
			t.Error(err)
		}
	})

	// 4. 开始验证字段
	if msgType != "room" {
		t.Fatalf(
			"expected type room, got %s",
			msgType,
		)
	}

	if fromID != 1 {
		t.Fatalf(
			"expected from_id 1, got %d",
			fromID,
		)
	}

	if fromNickname != "Tom" {
		t.Fatalf(
			"expected nickname Tom, got %s",
			fromNickname,
		)
	}

	// 房间消息没有 to_id
	if toID.Valid {
		t.Fatalf(
			"expected to_id NULL, got %d",
			toID.Int64,
		)
	}

	// room_id 应该有值
	if !roomID.Valid {
		t.Fatal("expected room_id not NULL")
	}

	if roomID.Int64 != 1 {
		t.Fatalf(
			"expected room_id 1, got %d",
			roomID.Int64,
		)
	}

	// room_name 应该有值
	if !roomName.Valid {
		t.Fatal("expected room_name not NULL")
	}

	if roomName.String != "golang" {
		t.Fatalf(
			"expected room_name golang, got %s",
			roomName.String,
		)
	}

	if gotContent != content {
		t.Fatalf(
			"expected content %s, got %s",
			content,
			gotContent,
		)
	}
}

func TestMessageServicePersistPrivate(t *testing.T) {
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
	messageService := NewMessageService(messageRepo)

	content := "test-private-message"

	// 先清理旧测试数据
	if _, err := db.Exec(
		"DELETE FROM messages WHERE content = ?",
		content,
	); err != nil {
		t.Fatal(err)
	}

	msg := domain.Message{
		ID:           rand.Text(),
		Type:         domain.MessagePrivate,
		From:         1,
		FromNickname: "Tom",
		To:           2,
		Content:      content,
		Time:         time.Now(),
	}

	// 1. 持久化
	if err := messageService.Persist(
		context.Background(),
		msg,
	); err != nil {
		t.Fatal(err)
	}

	// 2. 从数据库查回来
	var (
		id           int64
		msgType      string
		fromID       int64
		fromNickname string

		toID     sql.NullInt64
		roomID   sql.NullInt64
		roomName sql.NullString

		gotContent string
	)

	err = db.QueryRow(
		`SELECT
			id,
			type,
			from_id,
			from_nickname,
			to_id,
			room_id,
			room_name,
			content
		FROM messages
		WHERE content = ?
		ORDER BY id DESC
		LIMIT 1`,
		content,
	).Scan(
		&id,
		&msgType,
		&fromID,
		&fromNickname,
		&toID,
		&roomID,
		&roomName,
		&gotContent,
	)

	if err != nil {
		t.Fatal(err)
	}

	// 3. 测试结束后删除这条消息
	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM messages WHERE id = ?",
			id,
		)
		if err != nil {
			t.Error(err)
		}
	})

	// 4. 开始验证字段
	if msgType != "private" {
		t.Fatalf(
			"expected type private, got %s",
			msgType,
		)
	}

	if fromID != 1 {
		t.Fatalf(
			"expected from_id 1, got %d",
			fromID,
		)
	}

	if fromNickname != "Tom" {
		t.Fatalf(
			"expected nickname Tom, got %s",
			fromNickname,
		)
	}

	// 私聊消息必须有 to_id
	if !toID.Valid {
		t.Fatal("expected to_id not NULL")
	}

	if toID.Int64 != 2 {
		t.Fatalf(
			"expected to_id 2, got %d",
			toID.Int64,
		)
	}

	// 私聊消息没有 room_id
	if roomID.Valid {
		t.Fatalf(
			"expected room_id NULL, got %d",
			roomID.Int64,
		)
	}

	// 私聊消息没有 room_name
	if roomName.Valid {
		t.Fatalf(
			"expected room_name NULL, got %s",
			roomName.String,
		)
	}

	if gotContent != content {
		t.Fatalf(
			"expected content %s, got %s",
			content,
			gotContent,
		)
	}
}
