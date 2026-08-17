package model


type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Nickname     string
}


type Room struct {
	ID int64
	Name string
}