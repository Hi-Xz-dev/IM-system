package server

type RoomInfo struct {
	ID    int64
	Name  string
	Count int
}

type OnlineUser struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
}
