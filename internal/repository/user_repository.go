package repository

import (
	"IM-system/internal/model"
	"context"
	"database/sql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// 注册时把用户写进数据库
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	//执行 INSERT MySQL 自动生成 id
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (username, password_hash, nickname)
		 VALUES (?, ?, ?)`,
		user.Username,
		user.PasswordHash,
		user.Nickname,
	)
	if err != nil {
		return err
	}
	//取出ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id

	return nil
}

// → 登录时根据 username 查用户
// → 取出 password_hash 做密码校验
func (r *UserRepository) FindByUsername(ctx context.Context,username string,) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, nickname
		 FROM users
		 WHERE username = ?`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Nickname,		
	)
	if err != nil{
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(
	ctx context.Context,
	id int64,
)(*model.User, error){

	user := &model.User{}

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT id, username, password_hash, nickname
		FROM users
		WHERE id = ?
		`,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Nickname,
	)

	if err != nil{
		return nil, err
	}

	return user, nil
}
