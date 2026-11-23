package server

import (
	"context"
	"errors"

	"IM-system/internal/message"
	"IM-system/internal/model"
	"IM-system/room"
	"IM-system/user"
)

var (
	ErrRoomNotFound  = errors.New("room not found")
	ErrNotRoomMember = errors.New("not a room member")
)

// roomForMemberLocked checks account membership. The caller holds mapLock.
func (s *Server) roomForMemberLocked(userID, roomID int64) (*room.Room, error) {
	r, ok := s.Rooms[roomID]
	if !ok {
		return nil, ErrRoomNotFound
	}
	if len(r.Users[userID]) == 0 {
		return nil, ErrNotRoomMember
	}
	return r, nil
}

func (s *Server) GetMembers(userID, roomID int64) ([]OnlineUser, error) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	r, err := s.roomForMemberLocked(userID, roomID)
	if err != nil {
		return nil, err
	}
	members := make([]OnlineUser, 0, len(r.Users))
	for id := range r.Users {
		if profile := s.Profiles[id]; profile != nil {
			members = append(members, OnlineUser{ID: id, Nickname: profile.Nickname})
		}
	}
	return members, nil
}

func (s *Server) checkRoomMembership(userID, roomID int64) error {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	_, err := s.roomForMemberLocked(userID, roomID)
	return err
}

func (s *Server) GetRoomHistoryForUser(ctx context.Context, userID, roomID int64, before *model.HistoryPosition) (message.HistoryPage, error) {
	if err := s.checkRoomMembership(userID, roomID); err != nil {
		return message.HistoryPage{}, err
	}
	page, err := s.messageService.GetRoomHistory(ctx, roomID, before)
	if err != nil {
		return message.HistoryPage{}, err
	}
	// The caller may have left while the database query was running.
	if err := s.checkRoomMembership(userID, roomID); err != nil {
		return message.HistoryPage{}, err
	}
	return page, nil
}

func (s *Server) sendRoomAccessError(usr *user.User, err error) error {
	if err == ErrRoomNotFound {
		return s.SendSystemMessage(usr, "房间不存在")
	}
	return s.SendSystemMessage(usr, "当前未加入房间")
}
