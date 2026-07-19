package auth

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTService struct {
	secret string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret: secret,
	}
}

func (j *JWTService) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWTService) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (any, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}

			return []byte(j.secret), nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
	)
	if err != nil {
		return 0, err
	}
	//保存的是Payload 断言成jwt.MapClaims使用
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	userIDValue, ok := claims["user_id"]
	if !ok {
		return 0, errors.New("user-id not found in token")
	}

	userIDFloat, ok := userIDValue.(float64)
	if !ok {
		return 0, errors.New("invalid user_id type")
	}

	return int64(userIDFloat), nil
}
