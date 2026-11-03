package repository

import (
	"IM-system/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

var ErrRoomAlreadyExists = errors.New("room already exists")

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{
		db: db,
	}
}

func (r *RoomRepository) Create(
	ctx context.Context,
	room *model.Room,
) error {

	result, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO rooms (name)
		VALUES (?)
		`,
		room.Name,
	)

	if err != nil {

		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrRoomAlreadyExists
		}
		return fmt.Errorf("insert room: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return fmt.Errorf("get room insert id: %w", err)
	}

	room.ID = id

	return nil
}

// FindAll 加载房间资料，不包含在线成员。
func (r *RoomRepository) FindAll(ctx context.Context) ([]model.Room, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM rooms ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("query rooms: %w", err)
	}
	defer rows.Close()
	rooms := make([]model.Room, 0)
	for rows.Next() {
		var room model.Room
		if err := rows.Scan(&room.ID, &room.Name); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}
	return rooms, nil
}
