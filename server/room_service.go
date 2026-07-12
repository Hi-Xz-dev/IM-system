package server

import (
	"IM-system/room"
	"IM-system/user"
	"fmt"
	"strings"
)

// 房间人数
func (s *Server) Members(usr *user.User, roomName string) {
	s.mapLock.RLock()
	if !usr.InRoom(roomName){
		s.mapLock.RUnlock()
		usr.SendMsg("当前未加入房间")
		return
	}
	r, ok := s.Rooms[roomName]
	if !ok {
		s.mapLock.RUnlock()
		usr.SendMsg("房间不存在")
		return
	}
	names := make([]string, 0, len(r.Users))
	for name := range r.Users { //只要名字
		names = append(names, name)
	}
	s.mapLock.RUnlock()
	usr.SendMsg("房间成员：" + strings.Join(names, ","))	
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
		user.SendMsg(msg)
	}
}

// 加入房间
func (s *Server) JoinRoom(joinuser *user.User, roomName string) {
	s.mapLock.Lock()
	r, ok := s.Rooms[roomName]
	if !ok {
		s.mapLock.Unlock() //提前返回
		joinuser.SendMsg("房间名不存在，请重试")
		return //有返回 要么用defer 或者提前解锁
	}
	if joinuser.InRoom(roomName){
		s.mapLock.Unlock()
		joinuser.SendMsg("已加入房间" + roomName)
		return
	}
	users := make([]*user.User, 0, len(r.Users))
	for _, u := range r.Users {
		users = append(users, u)
	}
	s.joinRoomUnsafe(joinuser, roomName)
	s.mapLock.Unlock() //正常返回
	//广播消息
	msg := "[系统][" + joinuser.Name + "]" + "加入房间：" + roomName + "\n"
	s.RoomBroadcast(users, msg)
	joinuser.SendMsg("成功加入房间：" + roomName)
}

// 加入房间 (内层)
func (s *Server) joinRoomUnsafe(joinuser *user.User, roomName string) {
	r := s.Rooms[roomName]
	r.Users[joinuser.Name] = joinuser
	joinuser.AddRoom(roomName)

}

// 退出房间（外层）
func (s *Server) LeaveRoom(leaveuser *user.User, roomName string) {
	s.mapLock.Lock()
	if !leaveuser.InRoom(roomName) {
		s.mapLock.Unlock()
		leaveuser.SendMsg("当前未加入房间" + roomName)
		return
	}
	r, ok := s.Rooms[roomName]
	if !ok {
		s.mapLock.Unlock()
		leaveuser.SendMsg("当前房间不存在") //防御式编程（Defensive Programming）
		return
	}
	//make([]T, len, cap)
	users := make([]*user.User, 0, len(r.Users)) // snapshot
	for _, u := range r.Users {
		if u != leaveuser {
			users = append(users, u)
		}
	}
	//modify data
	s.leaveRoomUnsafe(leaveuser, roomName)
	s.mapLock.Unlock()
	msg := "[系统][" + leaveuser.Name + "]" + "离开房间：" + roomName + "\n"
	s.RoomBroadcast(users, msg)
	leaveuser.SendMsg("退出房间成功")
}

// 退出房间（内层）
func (s *Server) leaveRoomUnsafe(leaveuser *user.User, roomName string) {

	r, ok := s.Rooms[roomName]
	if !ok{
		leaveuser.RemoveRoom(roomName)
		return
	}
	delete(r.Users, leaveuser.Name)
	leaveuser.RemoveRoom(roomName)
	if len(r.Users) == 0 {
		delete(s.Rooms, roomName) //房间无成员直接删除房间
	}
}

// 创建房间
func (s *Server) CreateRoom(createuser *user.User, roomName string) {
	s.mapLock.Lock()
	if _, ok := s.Rooms[roomName]; ok {
		s.mapLock.Unlock()
		createuser.SendMsg("房间名存在，请重试")
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
func (s *Server) RoomChat(sender *user.User, roomName,content string) {
	s.mapLock.RLock()

	if !sender.InRoom(roomName) {
		s.mapLock.RUnlock()
		sender.SendMsg("当前用户未加入该房间")
		return
	}
	r , ok := s.Rooms[roomName]
	if !ok {
		s.mapLock.RUnlock()
		sender.SendMsg("当前房间不存在")
		return
	}
	// snapshot
	users := make([]*user.User, 0, len(r.Users))
	for _, u := range r.Users {
		users = append(users, u)
	}
	s.mapLock.RUnlock()
	msg := "[" + roomName + "][" + sender.Name + "] " + content
	s.RoomBroadcast(users, msg)
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
func (s *Server) GetMembers(roomName string) ([]string, bool) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	r, ok := s.Rooms[roomName]
	if !ok {
		return nil, false
	}
	roomUsers := make([]string, 0, len(r.Users))
	for name := range r.Users {
		roomUsers = append(roomUsers, name)
	}
	return roomUsers, true
}
//================HTTP========================

// 加入房间
func (s *Server) JoinRoomByName(usr, roomName string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	u, ok := s.OnlineUsers[usr]
	if !ok {
		return "未找到用户", false
	}
	if _, ok := s.Rooms[roomName]; !ok {
		return "未找到房间", false
	}
	if u.InRoom(roomName){
		return "已加入房间", true
	}
	s.joinRoomUnsafe(u, roomName)
	return "加入房间", true
}

// 退出房间
func (s *Server) LeaveRoomByName(usr, roomName string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	u, ok := s.OnlineUsers[usr]
	if !ok {
		return "未找到用户", false
	}
	if !u.InRoom(roomName){
		return "未加入房间", true
	}
	if _, ok := s.Rooms[roomName]; !ok{
		u.RemoveRoom(roomName)
		return "当前房间不存在", false
	}
	s.leaveRoomUnsafe(u, roomName)
	return "退出房间", true
}

//创建房间
func (s *Server)CreateRoomByName(userName, roomName string ) (string, bool){
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	createUser, ok := s.OnlineUsers[userName]
	if !ok {
		return "未找到用户", false
	}
	if  _, ok := s.Rooms[roomName]; ok {
		return "房间名存在，请重试",false
	}
	s.Rooms[roomName] = room.NewRoom(roomName)
	s.joinRoomUnsafe(createUser, roomName)
	return "创建房间成功", true
}	
