package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// 密码加密
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// 再次加密 与数据库Hash比对
func CheckPassword(passwordHash, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)
}
