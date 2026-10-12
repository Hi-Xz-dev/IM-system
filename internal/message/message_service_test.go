package message

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"IM-system/internal/domain"
	"IM-system/internal/repository"

	_ "github.com/go-sql-driver/mysql"
)

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
