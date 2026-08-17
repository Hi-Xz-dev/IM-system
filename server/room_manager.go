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
		s.mapLock.Unlock()
		s.SendSystemMessage(
			joinuser,
			"已加入房间"+r.Name,
		)
		return
	}
	s.joinRoomUnsafe(joinuser, roomID)

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
func (s *Server) CreateRoom(createuser *user.User, roomName string) error {

	modelRoom := &model.Room{
		Name: roomName,
	}

	if err := s.roomRepo.Create(context.Background(), modelRoom); err != nil {
		return err
	}

	roomID := modelRoom.ID

	s.mapLock.Lock()
	defer s.mapLock.Unlock()

	r := room.NewRoom(roomID, roomName)

	s.Rooms[roomID] = r

	s.joinRoomUnsafe(createuser, roomID)

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
	r := s.Rooms[roomID]
	r.Users[joinuser.ID] = append(r.Users[joinuser.ID], joinuser)
	joinuser.AddRoom(roomID)

}

// 退出房间
func (s *Server) leaveRoomUnsafe(leaveuser *user.User, roomID int64) {

	r, ok := s.Rooms[roomID]
	if !ok {
		leaveuser.RemoveRoom(roomID)
		return
	}
	delete(r.Users, leaveuser.ID)
	leaveuser.RemoveRoom(roomID)
	if len(r.Users) == 0 {
		delete(s.Rooms, roomID) //房间无成员直接删除房间
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
	for _, u := range users {
		if u.InRoom(roomID) {
			return "已加入房间", true
		}
	}
	for _, u := range users {
		s.joinRoomUnsafe(u, roomID)
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
			s.leaveRoomUnsafe(users[0], roomID)
			return "退出房间", true
		}
	}

	return "未加入房间", true
}

// 创建房间
func (s *Server) CreateRoomRuntime(userID int64, roomID int64, roomName string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	createUsers, ok := s.OnlineUsers[userID]
	if !ok {
		return "未找到用户", false
	}

	if _, ok := s.Rooms[roomID]; ok {
		return "房间名存在，请重试", false
	}

	r := room.NewRoom(roomID, roomName)

	s.Rooms[roomID] = r
	for _, u := range createUsers {
		s.joinRoomUnsafe(
			u,
			roomID,
		)
	}
	return "创建房间成功", true
}
