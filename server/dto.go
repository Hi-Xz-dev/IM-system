package server

type RoomInfo struct {
	Name  string
	Count int
}

type OnlineUser struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
}
