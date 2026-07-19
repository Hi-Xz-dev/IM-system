package auth

import(
	"golang.org/x/crypto/bcrypt"
)
//密码加密
func HashPassword(password string)(string, error){
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil{
		return "", err
	}
	return string(hash), nil
}
// 把 bcrypt 的结果返回出来不做判断
func CheckPassword(passwordHash, password string) error{
	return bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)
}