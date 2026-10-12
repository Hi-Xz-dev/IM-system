package repository

import (
	"IM-system/internal/model"

	"context"
	"database/sql"
	"fmt"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{
		db: db,
	}
}
//持久化
func (m *MessageRepository) Create(
	ctx context.Context,
	msg *model.Message,
) error {

	result, err := m.db.ExecContext(
		ctx,
		`INSERT INTO messages (
			type,
			from_id,
			from_nickname,
			to_id,
			room_id,
			room_name,
			content,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
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
		ORDER BY created_at DESC
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
		ORDER BY created_at DESC
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