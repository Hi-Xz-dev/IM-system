package auth

import (
	"IM-system/internal/config"
	"IM-system/internal/database"
	"IM-system/internal/repository"

	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func setupAuthService(t *testing.T) (*Service, *repository.UserRepository, *sql.DB) {
	cfg := config.MySQLConfig{
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "",
		DataName: "im_system",
	}

	db, err := database.NewMySQL(cfg)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := repository.NewUserRepository(db)

	jwtService := NewJWTService(
		"test-secret",
	)

	service := NewService(
		repo,
		jwtService,
	)

	return service, repo, db
}
func TestRegister(t *testing.T) {

	service, repo, db := setupAuthService(t)

	ctx := context.Background()

	username := "test_user"
	// 测试结束后清理数据
	t.Cleanup(func() {

		_, err := db.Exec(
			"DELETE FROM users WHERE username=?",
			username,
		)

		if err != nil {
			t.Fatal(err)
		}
	})

	// 清理旧数据
	if _, err := db.Exec(
		"DELETE FROM users WHERE username=?",
		username,
	); err != nil {
		t.Fatal(err)
	}

	err := service.Register(
		ctx,
		RegisterInput{
			Username: username,
			Password: "123456",
			Nickname: "Tom",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	user, err := repo.FindByUsername(
		ctx,
		username,
	)

	if err != nil {
		t.Fatal(err)
	}

	if user.Username != username {
		t.Fatal("username mismatch")
	}

	if user.ID == 0 {
		t.Fatal("user id not generated")
	}

	if user.PasswordHash == "123456" {
		t.Fatal("password stored plaintext")
	}

	err = CheckPassword(
		user.PasswordHash,
		"123456",
	)

	if err != nil {
		t.Fatal("password verify failed")
	}

}

func TestRegisteDupicate(t *testing.T) {

	service, _, db := setupAuthService(t)

	ctx := context.Background()

	username := "duplicate_user"

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM users WHERE username=?",
			username,
		)

		if err != nil {
			t.Fatal(err)
		}
	})

	// 测试开始清理旧数据
	_, err := db.Exec(
		"DELETE FROM users WHERE username=?",
		username,
	)

	if err != nil {
		t.Fatal(err)
	}

	//第一次注册
	err = service.Register(
		ctx,
		RegisterInput{
			Username: username,
			Password: "123456",
			Nickname: "Tom",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	//第二次注册相同用户名
	err = service.Register(
		ctx,
		RegisterInput{
			Username: username,
			Password: "123456",
			Nickname: "Tom",
		},
	)
	if err == nil {
		t.Fatal(err)
	}

	if !errors.Is(err, ErrUserAlreadyExists) {

		t.Fatalf(
			"expected ErrUserAlreadyExists, got %v",
			err,
		)
	}
}

func TestLogin(t *testing.T) {

	service, _, db := setupAuthService(t)

	ctx := context.Background()

	username := "test_user"
	// 测试结束后清理数据
	t.Cleanup(func() {

		_, err := db.Exec(
			"DELETE FROM users WHERE username=?",
			username,
		)

		if err != nil {
			t.Fatal(err)
		}
	})

	// 清理旧数据
	if _, err := db.Exec(
		"DELETE FROM users WHERE username=?",
		username,
	); err != nil {
		t.Fatal(err)
	}

	err := service.Register(
		ctx,
		RegisterInput{
			Username: username,
			Password: "123456",
			Nickname: "Tom",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := service.Login(
		ctx,
		LoginInput{
			Username: username,
			Password: "123456",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.Token == "" {
		t.Fatal("token is empty")
	}

	if result.UserID == 0 {
		t.Fatal("user id is empty")
	}

	if result.Username != username {
		t.Fatal("username mismatch")
	}

	if nickname, err := service.UpdateNickname(ctx, result.UserID, "  Jerry  "); err != nil || nickname != "Jerry" {
		t.Fatalf("update nickname = %q, %v", nickname, err)
	}
	relogged, err := service.Login(ctx, LoginInput{Username: username, Password: "123456"})
	if err != nil {
		t.Fatal(err)
	}
	if relogged.Nickname != "Jerry" {
		t.Fatalf("expected persisted nickname Jerry, got %q", relogged.Nickname)
	}
}

func TestLoginWrongPassword(t *testing.T) {

	service, _, db := setupAuthService(t)

	ctx := context.Background()

	username := "wrong_password_user"

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM users WHERE username=?",
			username,
		)

		if err != nil {
			t.Fatal(err)
		}
	})

	// 清理旧数据
	if _, err := db.Exec(
		"DELETE FROM users WHERE username=?",
		username,
	); err != nil {
		t.Fatal(err)
	}

	// 创建用户
	err := service.Register(
		ctx,
		RegisterInput{
			Username: username,
			Password: "123456",
			Nickname: "Tom",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	_, err = service.Login(
		ctx,
		LoginInput{
			Username: username,
			Password: "12345678",
		},
	)

	if err == nil {
		t.Fatal("expected login failure, but got nil")
	}

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

}

func TestAuthenticate(t *testing.T) {

	service, _, db := setupAuthService(t)

	ctx := context.Background()

	username := "test_user"

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM users WHERE username=?",
			username,
		)

		if err != nil {
			t.Fatal(err)
		}
	})

	// 清理旧数据
	if _, err := db.Exec(
		"DELETE FROM users WHERE username=?",
		username,
	); err != nil {
		t.Fatal(err)
	}

	// 创建用户
	err := service.Register(
		ctx,
		RegisterInput{
			Username: username,
			Password: "123456",
			Nickname: "Tom",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	result, err := service.Login(
		ctx,
		LoginInput{
			Username: username,
			Password: "123456",
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	user, err := service.Authenticate(
		ctx,
		result.Token,
	)

	if err != nil {
		t.Fatal(err)
	}

	if user.Username != username {
		t.Fatal("username mismatch")
	}

	if user.ID == 0 {
		t.Fatal("user id empty")
	}

	// 占住连接池的唯一连接，验证认证等待连接时也受查询期限约束。
	db.SetMaxOpenConns(1)
	held, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()

	outerCtx, cancel := context.WithTimeout(ctx, authLookupTimeout+3*time.Second)
	defer cancel()
	if _, err := service.Authenticate(outerCtx, result.Token); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected authentication query timeout, got %v", err)
	}
	if outerCtx.Err() != nil {
		t.Fatal("authentication waited for the outer deadline")
	}
}
