package server

import (
	"IM-system/internal/logger"
	"IM-system/internal/model"
	"IM-system/room"
	"IM-system/user"
	"context"

	"fmt"
	"strings"
)

// 房间人数
func (s *Server) Members(usr *user.User, roomID int64) {
	s.mapLock.RLock()

	if !usr.InRoom(roomID) {
		s.mapLock.RUnlock()
		_ = s.SendSystemMessage(usr, "当前未加入房间")
		return
	}

	r, ok := s.Rooms[roomID]

	if !ok {
		s.mapLock.RUnlock()
		_ = s.SendSystemMessage(usr, "房间不存在")
		return
	}

	names := make([]string, 0, len(r.Users))

	for userID := range r.Users {
		profile, ok := s.Profiles[userID]
		if !ok {
			continue
		}

		names = append(
			names,
			profile.Nickname,
		)
	}

	s.mapLock.RUnlock()

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
	messages, err := s.messageService.GetRoomHistory(
		context.Background(),
		roomID,
		20,
	)

	if err != nil {
		logger.Log.Error(
			"get room history failed",
			"user_id", usr.ID,
			"room_id", roomID,
			"error", err,
		)

		_ = s.SendSystemMessage(
			usr,
			"获取房间历史消息失败",
		)
		return
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		text := fmt.Sprintf(
			"[%s] %s: %s",
			msg.Time.Format("15:04:05"),
			msg.FromNickname,
			msg.Content,
		)

		_ = s.SendSystemMessage(
			usr,
			text,
		)
	}
}

// 加入房间
func (s *Server) JoinRoom(joinuser *user.User, roomID int64) {
	s.mapLock.Lock()

	r, ok := s.Rooms[roomID]

	if !ok {
		s.mapLock.Unlock()
		s.SendSystemMessage(
			joinuser,
			"房间名不存在，请重试",
		)
		return
	}
	if joinuser.InRoom(roomID) {
		s.joinAccountRoomUnsafe(joinuser, roomID)
		s.mapLock.Unlock()
		s.SendSystemMessage(
			joinuser,
			"已加入房间"+r.Name,
		)
		return
	}
	s.joinAccountRoomUnsafe(joinuser, roomID)

	users := s.getRoomUsersUnsafe(roomID)

	s.mapLock.Unlock() //正常返回
	//广播消息
	if err := s.SendRoomMessage(
		joinuser,
		users,
		roomID,
		r.Name,
		"加入房间",
	); err != nil {
		logger.Log.Error(
			"send room message failed",
			"error",
			err,
		)
	}
	if err := s.SendSystemMessage(
		joinuser,
		"成功加入房间："+r.Name,
	); err != nil {

		logger.Log.Error(
			"send system message failed",
			"error",
			err,
		)
	}
}

// 退出房间
func (s *Server) LeaveRoom(leaveuser *user.User, roomID int64) {
	s.mapLock.Lock()

	r, ok := s.Rooms[roomID]
	if !ok {
		s.mapLock.Unlock()
		s.SendSystemMessage(
			leaveuser,
			"当前房间不存在",
		) //防御式编程（Defensive Programming）
		return
	}

	if !leaveuser.InRoom(roomID) {
		s.mapLock.Unlock()
		s.SendSystemMessage(
			leaveuser,
			"当前未加入房间",
		)
		return
	}

	s.leaveRoomUnsafe(leaveuser, roomID)

	users := s.getRoomUsersUnsafe(roomID)

	s.mapLock.Unlock()
	if err := s.SendRoomMessage(
		leaveuser,
		users,
		roomID,
		r.Name,
		"离开房间",
	); err != nil {
		logger.Log.Error(
			"send room message failed",
			"error",
			err,
		)
	}
	if err := s.SendSystemMessage(
		leaveuser,
		"成功退出房间："+r.Name,
	); err != nil {

		logger.Log.Error(
			"send system message failed",
			"error",
			err,
		)
	}
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

// 返回房间成员
func (s *Server) GetMembers(roomID int64) ([]*user.User, bool) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	r, ok := s.Rooms[roomID]
	if !ok {
		return nil, false
	}
	roomUsers := make([]*user.User, 0)

	for _, userList := range r.Users {

		for _, u := range userList {

			roomUsers = append(roomUsers, u)
		}
	}

	return roomUsers, true
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

// 加入房间
func (s *Server) JoinRoomH(userID int64, roomID int64) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	users, ok := s.OnlineUsers[userID]
	if !ok {
		return fmt.Sprintf("未找到用户(id=%d), 当前在线:%d人", userID, len(s.OnlineUsers)), false
	}
	if _, ok := s.Rooms[roomID]; !ok {
		return "未找到房间", false
	}
	alreadyJoined := len(s.Rooms[roomID].Users[userID]) > 0
	s.joinAccountRoomUnsafe(users[0], roomID)
	if alreadyJoined {
		return "已加入房间", true
	}
	return "加入房间", true
}

// 退出房间
func (s *Server) LeaveRoomH(userID int64, roomID int64) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	users, ok := s.OnlineUsers[userID]
	if !ok {
		return "未找到用户", false
	}
	if _, ok := s.Rooms[roomID]; !ok {

		return "当前房间不存在", false
	}
	for _, u := range users {
		if u.InRoom(roomID) {
			s.leaveRoomUnsafe(u, roomID)
			return "退出房间", true
		}
	}

	return "未加入房间", true
}
