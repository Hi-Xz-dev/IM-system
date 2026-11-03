package service

import (
	"context"

	"IM-system/internal/model"
	"IM-system/internal/repository"
	"IM-system/server"
)

type RoomService struct {
	server *server.Server

	roomRepo *repository.RoomRepository
}

func NewRoomService(
	s *server.Server,
	roomRepo *repository.RoomRepository,
) *RoomService {
	return &RoomService{
		server:   s,
		roomRepo: roomRepo,
	}
}

func (rs *RoomService) GetRooms() []server.RoomInfo {
	return rs.server.GetRooms()
}

func (rs *RoomService) CreateRoom(
	ctx context.Context,
	userID int64,
	name string,
) (*model.Room, error) {
	return rs.server.CreateRoomForUser(ctx, userID, name)
}
