package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"IM-system/internal/model"

	_ "github.com/go-sql-driver/mysql"
)

func TestMessageRepositoryCreate(t *testing.T) {
	db, err := sql.Open(
		"mysql",
		"root@tcp(127.0.0.1:3306)/im_system?parseTime=true",
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	repo := NewMessageRepository(db)

	roomID := int64(1)
	roomName := "golang"

	msg := &model.Message{
		Type:         "room",
		FromID:       1,
		FromNickname: "Tom",
		RoomID:       &roomID,
		RoomName:     &roomName,
		Content:      "hello",
		CreatedAt:    time.Now(),
	}

	err = repo.Create(
		context.Background(),
		msg,
	)
	if err != nil {
		t.Fatal(err)
	}

	if msg.ID == 0 {
		t.Fatal("message id not generated")
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM messages WHERE id = ?",
			msg.ID,
		)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestFindByRoomID(t *testing.T) {
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

	repo := NewMessageRepository(db)

	ctx := context.Background()

	const roomID1 int64 = 1
	const roomID2 int64 = 2

	roomName1 := "golang"
	roomName2 := "backend"

	content1 := "room-history-test-1"
	content2 := "room-history-test-2"
	content3 := "room-history-other"

	// 1. 先清理上一次测试可能留下的数据
	if _, err := db.Exec(
		`DELETE FROM messages
		 WHERE content IN (?, ?, ?)`,
		content1,
		content2,
		content3,
	); err != nil {
		t.Fatal(err)
	}

	// 2. 测试结束后再次清理
	t.Cleanup(func() {
		_, err := db.Exec(
			`DELETE FROM messages
			 WHERE content IN (?, ?, ?)`,
			content1,
			content2,
			content3,
		)
		if err != nil {
			t.Error(err)
		}
	})

	// 3. 准备 roomID=1 的第一条消息
	msg1 := &model.Message{
		Type:         "room",
		FromID:       5,
		FromNickname: "Tom",
		RoomID:       ptrInt64(roomID1),
		RoomName:     &roomName1,
		Content:      content1,
		CreatedAt:    time.Now(),
	}

	// roomID=1 的第二条消息
	msg2 := &model.Message{
		Type:         "room",
		FromID:       6,
		FromNickname: "Jerry",
		RoomID:       ptrInt64(roomID1),
		RoomName:     &roomName1,
		Content:      content2,
		CreatedAt:    time.Now().Add(time.Second),
	}

	// roomID=2，用来验证不会被查出来
	msg3 := &model.Message{
		Type:         "room",
		FromID:       7,
		FromNickname: "Alice",
		RoomID:       ptrInt64(roomID2),
		RoomName:     &roomName2,
		Content:      content3,
		CreatedAt:    time.Now().Add(2 * time.Second),
	}

	// 4. 直接通过 Repository 写入
	if err := repo.Create(ctx, msg1); err != nil {
		t.Fatal(err)
	}

	if err := repo.Create(ctx, msg2); err != nil {
		t.Fatal(err)
	}

	if err := repo.Create(ctx, msg3); err != nil {
		t.Fatal(err)
	}

	// 5. 查询 roomID=1 的历史消息
	messages, err := repo.FindByRoomID(
		ctx,
		roomID1,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	// 6. 应该只拿到 roomID=1 的两条
	if len(messages) != 2 {
		t.Fatalf(
			"expected 2 messages, got %d",
			len(messages),
		)
	}

	// 7. 确保没有混入 roomID=2
	for _, msg := range messages {
		if msg.RoomID == nil {
			t.Fatal("expected room_id not nil")
		}

		if *msg.RoomID != roomID1 {
			t.Fatalf(
				"expected room_id %d, got %d",
				roomID1,
				*msg.RoomID,
			)
		}
	}

	// 8. 因为 SQL 是 ORDER BY created_at DESC，从大到小(时间)
	// 第二条消息应该排在前面
	if messages[0].Content != content2 {
		t.Fatalf(
			"expected first message %s, got %s",
			content2,
			messages[0].Content,
		)
	}

	if messages[1].Content != content1 {
		t.Fatalf(
			"expected second message %s, got %s",
			content1,
			messages[1].Content,
		)
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}

func TestFindPrivateMessages(t *testing.T) {
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

	repo := NewMessageRepository(db)
	ctx := context.Background()

	content1 := "alice-to-bob-test"
	content2 := "bob-to-alice-test"
	content3 := "alice-to-other-test"

	// 1. 清理之前可能残留的测试数据
	if _, err := db.Exec(
		`DELETE FROM messages
		WHERE content IN (?, ?, ?)`,
		content1,
		content2,
		content3,
	); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			`DELETE FROM messages
			WHERE content IN (?, ?, ?)`,
			content1,
			content2,
			content3,
		)

		if err != nil {
			t.Error(err)
		}
	})

	// 2. Alice -> Bob
	toBob := int64(2)

	msg1 := &model.Message{
		Type:         "private",
		FromID:       1,
		FromNickname: "Alice",
		ToID:         &toBob,
		Content:      content1,
		CreatedAt:    time.Now(),
	}

	// 3. Bob -> Alice
	toAlice := int64(1)

	msg2 := &model.Message{
		Type:         "private",
		FromID:       2,
		FromNickname: "Bob",
		ToID:         &toAlice,
		Content:      content2,
		CreatedAt:    time.Now().Add(time.Second),
	}

	// 4. Alice -> 用户3，作为干扰数据
	toOther := int64(3)

	msg3 := &model.Message{
		Type:         "private",
		FromID:       1,
		FromNickname: "Alice",
		ToID:         &toOther,
		Content:      content3,
		CreatedAt:    time.Now().Add(2 * time.Second),
	}

	if err := repo.Create(ctx, msg1); err != nil {
		t.Fatal(err)
	}

	if err := repo.Create(ctx, msg2); err != nil {
		t.Fatal(err)
	}

	if err := repo.Create(ctx, msg3); err != nil {
		t.Fatal(err)
	}

	// 5. 查询 Alice 和 Bob 的聊天记录
	messages, err := repo.FindPrivateMessages(
		ctx,
		1,
		2,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}

	// 6. 应该只有两条
	if len(messages) != 2 {
		t.Fatalf(
			"expected 2 messages, got %d",
			len(messages),
		)
	}

	// 7. 最新的 Bob -> Alice 应该在前面
	if messages[0].Content != content2 {
		t.Fatalf(
			"expected first message %s, got %s",
			content2,
			messages[0].Content,
		)
	}

	// 8. Alice -> Bob 在后面
	if messages[1].Content != content1 {
		t.Fatalf(
			"expected second message %s, got %s",
			content1,
			messages[1].Content,
		)
	}
}