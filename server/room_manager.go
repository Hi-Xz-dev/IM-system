package server

import (
	"IM-system/internal/logger"
	"IM-system/room"
	"IM-system/user"

	"fmt"
	"strings"
)

// 房间人数
func (s *Server) Members(usr *user.User, roomName string) {
	s.mapLock.RLock()

	if !usr.InRoom(roomName) {
		s.mapLock.RUnlock()
		_ = s.SendSystemMessage(usr, "当前未加入房间")
		return
	}

	r, ok := s.Rooms[roomName]

	if !ok {
		s.mapLock.RUnlock()
		_ = s.SendSystemMessage(usr, "房间不存在")
		return
	}
	names := make(map[string]struct{})

	for _, users := range r.Users {

		for _, u := range users {

			names[u.Nickname] = struct{}{}
		}
	}

	s.mapLock.RUnlock()

	result := make([]string, 0, len(names))

	for name := range names {

		result = append(result, name)
	}
	_ = s.SendSystemMessage(usr, "房间成员："+strings.Join(result, ","))
}

// 显示房间
func (s *Server) ShowRooms(user *user.User) {
	s.mapLock.RLock()
	// snapshot：锁内构建消息
	msgs := make([]string, 0, len(s.Rooms))
	for name, r := range s.Rooms {
		msg := fmt.Sprintf("[房间]%s 人数：%d\n", name, len(r.Users))
		msgs = append(msgs, msg)
	}
	s.mapLock.RUnlock()
	// 锁外 IO
	for _, msg := range msgs {
		_ = s.SendSystemMessage(user, msg)
	}
}

// 加入房间
func (s *Server) JoinRoom(joinuser *user.User, roomName string) {
	s.mapLock.Lock()

	if _, ok := s.Rooms[roomName]; !ok {
		s.mapLock.Unlock()
		s.SendSystemMessage(
			joinuser,
			"房间名不存在，请重试",
		)
		return
	}
	if joinuser.InRoom(roomName) {
		s.mapLock.Unlock()
		s.SendSystemMessage(
			joinuser,
			"已加入房间 "+roomName,
		)
		return
	}
	s.joinRoomUnsafe(joinuser, roomName)

	users := s.getRoomUsersUnsafe(roomName)

	s.mapLock.Unlock() //正常返回
	//广播消息
	if err := s.SendRoomMessage(
		joinuser,
		users,
		roomName,
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
		"成功加入房间："+roomName,
	); err != nil {

		logger.Log.Error(
			"send system message failed",
			"error",
			err,
		)
	}
}

// 退出房间
func (s *Server) LeaveRoom(leaveuser *user.User, roomName string) {
	s.mapLock.Lock()

	if _, ok := s.Rooms[roomName]; !ok {
		s.mapLock.Unlock()
		s.SendSystemMessage(
			leaveuser,
			"当前房间不存在",
		) //防御式编程（Defensive Programming）
		return
	}

	if !leaveuser.InRoom(roomName) {
		s.mapLock.Unlock()
		s.SendSystemMessage(
			leaveuser,
			"当前未加入房间",
		)
		return
	}

	s.leaveRoomUnsafe(leaveuser, roomName)

	users := s.getRoomUsersUnsafe(roomName)

	s.mapLock.Unlock()
	if err := s.SendRoomMessage(
		leaveuser,
		users,
		roomName,
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
		"成功退出房间："+roomName,
	); err != nil {

		logger.Log.Error(
			"send system message failed",
			"error",
			err,
		)
	}
}

// 创建房间
func (s *Server) CreateRoom(createuser *user.User, roomName string) {
	s.mapLock.Lock()
	if _, ok := s.Rooms[roomName]; ok {
		s.mapLock.Unlock()
		_ = s.SendSystemMessage(createuser, "房间名存在，请重试")
		return
	}
	//1. 把新房间加入 Server.Rooms
	s.Rooms[roomName] = room.NewRoom(roomName)
	//2. 把创建者加入 Room.Users
	//3. 把房间加入 User.JoinedRooms
	s.joinRoomUnsafe(createuser, roomName)
	s.mapLock.Unlock()
}

// 群聊功能
func (s *Server) RoomChat(sender *user.User, roomName, content string) {
	s.mapLock.RLock()

	if _, ok := s.Rooms[roomName]; !ok {
		s.mapLock.RUnlock()
		s.SendSystemMessage(
			sender,
			"当前房间不存在",
		)
		return
	}

	if !sender.InRoom(roomName) {
		s.mapLock.RUnlock()
		s.SendSystemMessage(
			sender,
			"当前用户未加入房间",
		)
		return
	}

	users := s.getRoomUsersUnsafe(roomName)

	s.mapLock.RUnlock()

	if err := s.SendRoomMessage(
		sender,
		users,
		roomName,
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
	for name, r := range s.Rooms {
		info := RoomInfo{
			Name:  name,
			Count: len(r.Users),
		}
		rooms = append(rooms, info)
	}
	return rooms
}

// 返回房间成员
func (s *Server) GetMembers(roomName string) ([]*user.User, bool) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	r, ok := s.Rooms[roomName]
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

// ================Unsafe========================
// 加入房间 (内层)
func (s *Server) joinRoomUnsafe(joinuser *user.User, roomName string) {
	r := s.Rooms[roomName]
	r.Users[joinuser.ID] = append(r.Users[joinuser.ID], joinuser)
	joinuser.AddRoom(roomName)

}

// 退出房间
func (s *Server) leaveRoomUnsafe(leaveuser *user.User, roomName string) {

	r, ok := s.Rooms[roomName]
	if !ok {
		leaveuser.RemoveRoom(roomName)
		return
	}
	delete(r.Users, leaveuser.ID)
	leaveuser.RemoveRoom(roomName)
	if len(r.Users) == 0 {
		delete(s.Rooms, roomName) //房间无成员直接删除房间
	}
}

// 返回房间用户切片
func (s *Server) getRoomUsersUnsafe(
	roomName string,
) []*user.User {

	r, ok := s.Rooms[roomName]

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
func (s *Server) JoinRoomH(userID int64, roomName string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	users, ok := s.OnlineUsers[userID]
	if !ok {
		return fmt.Sprintf("未找到用户(id=%d), 当前在线:%d人", userID, len(s.OnlineUsers)), false
	}
	if _, ok := s.Rooms[roomName]; !ok {
		return "未找到房间", false
	}
	for _, u := range users {
		if u.InRoom(roomName) {
			return "已加入房间", true
		}
	}
	for _, u := range users {
		s.joinRoomUnsafe(u, roomName)
	}
	return "加入房间", true
}

// 退出房间
func (s *Server) LeaveRoomH(userID int64, roomName string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	users, ok := s.OnlineUsers[userID]
	if !ok {
		return "未找到用户", false
	}
	if _, ok := s.Rooms[roomName]; !ok {

		return "当前房间不存在", false
	}
	for _, u := range users {
		if u.InRoom(roomName) {
			s.leaveRoomUnsafe(users[0], roomName)
			return "退出房间", true
		}
	}

	return "未加入房间", true
}

// 创建房间
func (s *Server) CreateRoomH(userID int64, roomName string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	createUser, ok := s.OnlineUsers[userID]
	if !ok {
		return "未找到用户", false
	}
	if _, ok := s.Rooms[roomName]; ok {
		return "房间名存在，请重试", false
	}
	s.Rooms[roomName] = room.NewRoom(roomName)
	for _, u := range createUser {
		s.joinRoomUnsafe(u, roomName)
	}
	return "创建房间成功", true
}
