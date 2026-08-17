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

func NewRoomService(s *server.Server,roomRepo *repository.RoomRepository,) *RoomService {
	return &RoomService{
		server: s,
		roomRepo: roomRepo,
	}
}

func (rs *RoomService) GetRooms() []server.RoomInfo {
	return rs.server.GetRooms()
}

func (rs *RoomService) CreateRoomRecord(
	ctx context.Context,
	name string,
) (*model.Room, error) {

	modelRoom := &model.Room{
		Name: name,
	}

	if err := rs.roomRepo.Create(ctx, modelRoom); err != nil{
		return nil, err
	}


	return modelRoom, nil
}


