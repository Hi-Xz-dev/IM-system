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
	before *model.HistoryPosition,
) ([]model.Message, error) {
	return m.findHistory(ctx, "room_id = ?", []any{roomID}, limit, before)
}

func (m *MessageRepository) FindPrivateMessages(
	ctx context.Context,
	userID1, userID2 int64,
	limit int,
	before *model.HistoryPosition,
) ([]model.Message, error) {
	return m.findHistory(ctx, `type = 'private' AND (
		(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)
	)`, []any{userID1, userID2, userID2, userID1}, limit, before)
}

func (m *MessageRepository) findHistory(
	ctx context.Context,
	filter string,
	args []any,
	limit int,
	before *model.HistoryPosition,
) ([]model.Message, error) {
	// filter comes from the two fixed queries above. All values use placeholders.
	query := `SELECT id, message_id, type, from_id, from_nickname,
			to_id, room_id, room_name, content, created_at
		FROM messages WHERE ` + filter
	if before != nil {
		query += " AND (created_at < ? OR (created_at = ? AND id < ?))"
		args = append(args, before.CreatedAt, before.CreatedAt, before.ID)
	}
	query += " ORDER BY created_at DESC, id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := m.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query message history: %w", err)
	}
	defer rows.Close()

	messages := make([]model.Message, 0)
	for rows.Next() {
		var msg model.Message
		if err := rows.Scan(
			&msg.ID, &msg.MessageID, &msg.Type, &msg.FromID, &msg.FromNickname,
			&msg.ToID, &msg.RoomID, &msg.RoomName, &msg.Content, &msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan history message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate message history: %w", err)
	}
	return messages, nil
}

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
