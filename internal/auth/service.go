package auth

import (
	"IM-system/internal/model"
	"IM-system/internal/repository"
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"strings"
)

type Service struct {
	userRepo *repository.UserRepository
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
		userRepo: userRepo,
		jwtService: jwtService,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) error {
	input.Username = strings.TrimSpace(input.Username)
	input.Nickname = strings.TrimSpace(input.Nickname)

	if input.Username == "" {
		return ErrInvalidUsername
	}

	if input.Password == "" {
		return ErrInvalidPassword
	}

	if input.Nickname == "" {
		return ErrInvalidNickname
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     input.Username,
		PasswordHash: passwordHash,
		Nickname:     input.Nickname,
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

func (s *Service) Authenticate(ctx context.Context, token string)(*model.User, error){
	userID, err := s.jwtService.ParseToken(token)
	if err != nil{
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx,userID,)
	if err != nil{
		return nil, err
	}
	return user, nil
}