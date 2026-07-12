package service

import(
	"IM-system/server"
)

type RoomService struct {
	server *server.Server
}

func NewRoomService(s *server.Server) *RoomService {
	return &RoomService{
		server : s,
	}
}

func (rs *RoomService) GetRooms() []server.RoomInfo{
	return rs.server.GetRooms()
}