package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"IM-system/internal/model"

	"github.com/go-sql-driver/mysql"
)

type MessageRepository struct {
	db *sql.DB
}

var ErrDuplicateKey = errors.New("duplicate key")

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}

// 持久化
func (m *MessageRepository) Create(
	ctx context.Context,
	msg *model.Message,
) error {

	if msg.MessageID == "" {
		return fmt.Errorf("message ID is required")
	}

	result, err := m.db.ExecContext(
		ctx,
		`INSERT INTO messages (
			message_id,
			type,
			from_id,
			from_nickname,
			to_id,
			room_id,
			room_name,
			content,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.MessageID,
		msg.Type,
		msg.FromID,
		msg.FromNickname,
		msg.ToID,
		msg.RoomID,
		msg.RoomName,
		msg.Content,
		msg.CreatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return fmt.Errorf("%w: %w", ErrDuplicateKey, err)
		}

		return err
	}
	//取出ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	msg.ID = id

	return nil
}

func (m *MessageRepository) FindByRoomID(
	ctx context.Context,
	roomID int64,
	limit int,
) ([]model.Message, error) {
	rows, err := m.db.QueryContext(
		ctx,
		`SELECT
			id,
			message_id,
			type,
			from_id,
			from_nickname,
			to_id,
			room_id,
			room_name,
			content,
			created_at
		FROM messages
		WHERE room_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?`,
		roomID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query room messages: %w",
			err,
		)
	}
	defer rows.Close()

	messages := make([]model.Message, 0)

	for rows.Next() {
		var msg model.Message

		if err := rows.Scan(
			&msg.ID,
			&msg.MessageID,
			&msg.Type,
			&msg.FromID,
			&msg.FromNickname,
			&msg.ToID,
			&msg.RoomID,
			&msg.RoomName,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan room message: %w",
				err,
			)
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate room messages: %w",
			err,
		)
	}

	return messages, nil
}

func (r *MessageRepository) FindPrivateMessages(
	ctx context.Context,
	userID1 int64,
	userID2 int64,
	limit int,
) ([]model.Message, error) {

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			id,
			message_id,
			type,
			from_id,
			from_nickname,
			to_id,
			room_id,
			room_name,
			content,
			created_at
		FROM messages
		WHERE type = 'private'
		  AND (
			(from_id = ? AND to_id = ?)
			OR
			(from_id = ? AND to_id = ?)
		  )
		ORDER BY created_at DESC, id DESC
		LIMIT ?`,
		userID1,
		userID2,
		userID2,
		userID1,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query private messages: %w",
			err,
		)
	}
	defer rows.Close()

	messages := make([]model.Message, 0)

	for rows.Next() {
		var msg model.Message

		if err := rows.Scan(
			&msg.ID,
			&msg.MessageID,
			&msg.Type,
			&msg.FromID,
			&msg.FromNickname,
			&msg.ToID,
			&msg.RoomID,
			&msg.RoomName,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan private message: %w",
				err,
			)
		}

		messages = append(
			messages,
			msg,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate private messages: %w",
			err,
		)
	}

	return messages, nil
}

// 遇到重复时，根据 message_id 查询已有消息
func (m *MessageRepository) FindByMessageID(
	ctx context.Context,
	messageID string,
) (*model.Message, error) {
	var msg model.Message

	err := m.db.QueryRowContext(
		ctx,
		`SELECT
			id, message_id, type, from_id, from_nickname,
			to_id, room_id, room_name, content, created_at
		FROM messages
		WHERE message_id = ?`,
		messageID,
	).Scan(
		&msg.ID,
		&msg.MessageID,
		&msg.Type,
		&msg.FromID,
		&msg.FromNickname,
		&msg.ToID,
		&msg.RoomID,
		&msg.RoomName,
		&msg.Content,
		&msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find message by message ID: %w", err)
	}

	return &msg, nil
}
