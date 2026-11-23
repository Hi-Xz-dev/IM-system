package server

import (
	"IM-system/internal/domain"
	"IM-system/internal/logger"
	"IM-system/internal/model"
	"IM-system/internal/protocol"
	"IM-system/room"
	"IM-system/user"
	"context"

	"fmt"
	"strings"
)

// 房间人数
func (s *Server) Members(usr *user.User, roomID int64) {
	members, err := s.GetMembers(usr.ID, roomID)
	if err != nil {
		_ = s.sendRoomAccessError(usr, err)
		return
	}
	names := make([]string, 0, len(members))
	for _, member := range members {
		names = append(names, member.Nickname)
	}

	_ = s.SendSystemMessage(
		usr,
		"房间成员："+strings.Join(names, ","),
	)
}

// 显示房间
func (s *Server) ShowRooms(user *user.User) {
	s.mapLock.RLock()
	// snapshot：锁内构建消息
	msgs := make([]string, 0, len(s.Rooms))
	for name, r := range s.Rooms {
		msg := fmt.Sprintf("[房间]%d 人数：%d\n", name, len(r.Users))
		msgs = append(msgs, msg)
	}
	s.mapLock.RUnlock()
	// 锁外 IO
	for _, msg := range msgs {
		_ = s.SendSystemMessage(user, msg)
	}
}

// 房间历史信息
func (s *Server) RoomHistory(
	usr *user.User,
	roomID int64,
) {
	page, err := s.GetRoomHistoryForUser(context.Background(), usr.ID, roomID, nil)
	if err != nil {
		switch err {
		case ErrRoomNotFound:
			s.sendCommandError(usr, domain.CmdRoomHistory, "room_not_found", "房间不存在")
			return
		case ErrNotRoomMember:
			s.sendCommandError(usr, domain.CmdRoomHistory, "not_room_member", "当前未加入房间")
			return
		}
		logger.Log.Error(
			"get room history failed",
			"user_id", usr.ID,
			"room_id", roomID,
			"error", err,
		)

		s.sendCommandError(usr, domain.CmdRoomHistory, "history_query_failed", "获取房间历史消息失败")
		return
	}

	s.sendHistory(usr, protocol.HistoryResponse{RoomID: roomID, Messages: page.Messages})
}

// 加入房间
func (s *Server) JoinRoom(usr *user.User, roomID int64) {
	result, _ := s.joinRoomForUser(usr.ID, roomID)
	_ = s.SendSystemMessage(usr, result)
}

// joinRoomForUser 负责账号入房，首次加入时沿用房间消息的发布和投递流程。
func (s *Server) joinRoomForUser(userID, roomID int64) (string, bool) {
	s.mapLock.Lock()
	r, ok := s.Rooms[roomID]
	if !ok {
		s.mapLock.Unlock()
		return "未找到房间", false
	}
	sessions := s.OnlineUsers[userID]
	if len(sessions) == 0 {
		s.mapLock.Unlock()
		return "未找到在线用户", false
	}

	sender := sessions[0]
	roomName := r.Name
	alreadyJoined := len(r.Users[userID]) > 0
	s.joinAccountRoomUnsafe(sender, roomID)
	if alreadyJoined {
		s.mapLock.Unlock()
		return "已加入房间：" + roomName, true
	}

	recipients := s.getRoomUsersUnsafe(roomID)
	s.mapLock.Unlock()

	// 成员状态已更新。发布通知失败时记录日志，入房仍然成功。
	if err := s.SendRoomMessage(sender, recipients, roomID, roomName, "加入房间"); err != nil {
		logger.Log.Error("send room join notification failed", "user_id", userID, "room_id", roomID, "error", err)
	}
	return "成功加入房间：" + roomName, true
}

// 退出房间
func (s *Server) LeaveRoom(usr *user.User, roomID int64) {
	result, _ := s.leaveRoomForUser(usr.ID, roomID)
	_ = s.SendSystemMessage(usr, result)
}

// leaveRoomForUser 处理账号主动退房；普通断线仍由 Offline 清理单条连接。
func (s *Server) leaveRoomForUser(userID, roomID int64) (string, bool) {
	s.mapLock.Lock()
	r, ok := s.Rooms[roomID]
	if !ok {
		s.mapLock.Unlock()
		return "未找到房间", false
	}
	if len(s.OnlineUsers[userID]) == 0 {
		s.mapLock.Unlock()
		return "未找到在线用户", false
	}
	members := r.Users[userID]
	if len(members) == 0 {
		s.mapLock.Unlock()
		return "未加入房间", true
	}

	sender := members[0]
	roomName := r.Name
	s.leaveRoomUnsafe(sender, roomID)
	recipients := s.getRoomUsersUnsafe(roomID)
	s.mapLock.Unlock()

	if err := s.SendRoomMessage(sender, recipients, roomID, roomName, "离开房间"); err != nil {
		logger.Log.Error("send room leave notification failed", "user_id", userID, "room_id", roomID, "error", err)
	}
	return "成功退出房间：" + roomName, true
}

// 创建房间
func (s *Server) CreateRoom(createuser *user.User, roomName string) (int64, error) {
	created, err := s.CreateRoomForUser(context.Background(), createuser.ID, roomName)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

// CreateRoomForUser 先持久化房间，再发布运行时状态；账号不在线时创建空房。
func (s *Server) CreateRoomForUser(ctx context.Context, userID int64, name string) (*model.Room, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("invalid room name")
	}
	created := &model.Room{Name: name}
	if err := s.roomRepo.Create(ctx, created); err != nil {
		return nil, err
	}
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	s.Rooms[created.ID] = room.NewRoom(created.ID, created.Name)
	if sessions := s.OnlineUsers[userID]; len(sessions) > 0 {
		s.joinAccountRoomUnsafe(sessions[0], created.ID)
	}
	return created, nil
}

// LoadRooms 仅在开放网络入口前调用，不恢复离线成员，也不覆盖已有在线状态。
func (s *Server) LoadRooms(ctx context.Context) error {
	rooms, err := s.roomRepo.FindAll(ctx)
	if err != nil {
		return err
	}
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	for _, saved := range rooms {
		if _, exists := s.Rooms[saved.ID]; !exists {
			s.Rooms[saved.ID] = room.NewRoom(saved.ID, saved.Name)
		}
	}
	return nil
}

// 群聊功能
func (s *Server) RoomChat(sender *user.User, roomID int64, content string) {
	s.mapLock.RLock()

	r, ok := s.Rooms[roomID]
	if !ok {
		s.mapLock.RUnlock()
		s.SendSystemMessage(
			sender,
			"当前房间不存在",
		)
		return
	}

	if !sender.InRoom(roomID) {
		s.mapLock.RUnlock()
		s.SendSystemMessage(
			sender,
			"当前用户未加入房间",
		)
		return
	}

	users := s.getRoomUsersUnsafe(roomID)

	s.mapLock.RUnlock()

	if err := s.SendRoomMessage(
		sender,
		users,
		roomID,
		r.Name,
		content,
	); err != nil {
		logger.Log.Error(
			"send room message failed",
			"error",
			err,
		)
	}

}

// 在线房间及人数
func (s *Server) GetRooms() []RoomInfo {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	// snapshot：锁内构建消息
	rooms := make([]RoomInfo, 0, len(s.Rooms))
	for id, r := range s.Rooms {
		info := RoomInfo{
			ID:    id,
			Name:  r.Name,
			Count: len(r.Users),
		}
		rooms = append(rooms, info)
	}
	return rooms
}

// 把一个已经创建好的 Room 加进 Server.Rooms
func (s *Server) AddRoom(r *room.Room) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()

	s.Rooms[r.ID] = r
}

// ================Unsafe========================
// 加入房间 (内层)
func (s *Server) joinRoomUnsafe(joinuser *user.User, roomID int64) {
	//连接是否已经存在
	r := s.Rooms[roomID]
	for _, session := range r.Users[joinuser.ID] {
		if session == joinuser {
			return
		}
	}
	r.Users[joinuser.ID] = append(r.Users[joinuser.ID], joinuser)
	joinuser.AddRoom(roomID)

}

// 加入操作以账号为单位，同步所有在线端。调用方必须持有 mapLock 写锁。
func (s *Server) joinAccountRoomUnsafe(usr *user.User, roomID int64) {
	s.joinRoomUnsafe(usr, roomID)
	for _, session := range s.OnlineUsers[usr.ID] {
		if !session.IsClosed {
			s.joinRoomUnsafe(session, roomID)
		}
	}
}

// 主动退房：同账号所有已加入的连接一起退出。调用方必须持有 mapLock 写锁。
func (s *Server) leaveRoomUnsafe(leaveuser *user.User, roomID int64) {

	r, ok := s.Rooms[roomID]
	if !ok {
		leaveuser.RemoveRoom(roomID)
		return
	}
	// 清理房间中这个账号所有连接的反向记录。
	for _, session := range r.Users[leaveuser.ID] {
		session.RemoveRoom(roomID)
	}

	leaveuser.RemoveRoom(roomID)
	// 删除整个账号在房间中的成员记录。
	delete(r.Users, leaveuser.ID)

}

// 断线清理：只移除当前连接，保留同账号其他连接。调用方必须持有 mapLock 写锁。
func (s *Server) removeRoomSessionUnsafe(usr *user.User, roomID int64) {
	usr.RemoveRoom(roomID)
	r, ok := s.Rooms[roomID]
	if !ok {
		return
	}

	sessions := r.Users[usr.ID]
	remaining := make([]*user.User, 0, len(sessions))
	for _, session := range sessions {
		if session != usr {
			remaining = append(remaining, session)
		}
	}
	if len(remaining) == 0 {
		delete(r.Users, usr.ID)
	} else {
		r.Users[usr.ID] = remaining
	}
}

// 返回房间用户切片
func (s *Server) getRoomUsersUnsafe(
	roomID int64,
) []*user.User {

	r, ok := s.Rooms[roomID]

	if !ok {
		return nil
	}

	users := make([]*user.User, 0)

	for _, sessions := range r.Users {

		for _, u := range sessions {

			users = append(users, u)
		}
	}

	return users
}

//================HTTP========================

// HTTP 加入房间
func (s *Server) JoinRoomH(userID int64, roomID int64) (string, bool) {
	return s.joinRoomForUser(userID, roomID)
}

// HTTP 退出房间
func (s *Server) LeaveRoomH(userID, roomID int64) (string, bool) {
	return s.leaveRoomForUser(userID, roomID)
}
