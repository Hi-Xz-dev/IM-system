package auth

import (
	"IM-system/internal/model"
	"IM-system/internal/repository"

	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
)

type Service struct {
	userRepo   *repository.UserRepository
	jwtService *JWTService
}

var (
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidNickname    = errors.New("invalid nickname")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type RegisterInput struct {
	Username string
	Password string
	Nickname string
}

type LoginInput struct {
	Username string
	Password string
}

type LoginResult struct {
	UserID   int64
	Username string
	Nickname string
	Token    string
}

func NewService(userRepo *repository.UserRepository, jwtService *JWTService) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// 注册
func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	input.Username = strings.TrimSpace(input.Username)

	if input.Username == "" {
		return ErrInvalidUsername
	}

	if input.Password == "" {
		return ErrInvalidPassword
	}

	nickname, err := normalizeNickname(input.Nickname)
	if err != nil {
		return err
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     input.Username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		var mysqlErr *mysql.MySQLError
		// 把一个 error 尝试转换成指定类型
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

func normalizeNickname(nickname string) (string, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" || !utf8.ValidString(nickname) || utf8.RuneCountInString(nickname) > 64 {
		return "", ErrInvalidNickname
	}
	return nickname, nil
}

func (s *Service) UpdateNickname(ctx context.Context, userID int64, nickname string) (string, error) {
	nickname, err := normalizeNickname(nickname)
	if err != nil {
		return "", err
	}
	if err := s.userRepo.UpdateNickname(ctx, userID, nickname); err != nil {
		return "", err
	}
	return nickname, nil
}

// 登录
func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	input.Username = strings.TrimSpace(input.Username)

	if input.Username == "" {
		return nil, ErrInvalidUsername
	}
	if input.Password == "" {
		return nil, ErrInvalidPassword
	}

	user, err := s.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	err = CheckPassword(
		user.PasswordHash,
		input.Password,
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(user.ID)

	return &LoginResult{
		UserID:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Token:    token,
	}, nil
}

// 拿身份证验证身份
func (s *Service) Authenticate(ctx context.Context, token string) (*model.User, error) {

	userID, err := s.jwtService.ParseToken(token)

	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
