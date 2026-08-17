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
