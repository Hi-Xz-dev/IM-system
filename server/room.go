package server

import (
	"IM-system/room"
	"IM-system/user"
	"fmt"
)

// 房间人数
func (s *Server) Members(u *user.User) {
	s.mapLock.RLock()
	roomName := u.CurrentRoom
	if roomName == "" {
		s.mapLock.RUnlock()
		u.SendMsg("当前未加入房间")
		return
	}
	r, ok := s.Rooms[roomName]
	if !ok {
		s.mapLock.RUnlock()
		u.SendMsg("房间不存在")
		return
	}
	members := make([]string, 0, len(r.Users))
	for name := range r.Users { //只要名字
		members = append(members, name)
	}
	s.mapLock.RUnlock()
	u.SendMsg("当前房间成员：")
	for _, name := range members {
		u.SendMsg(name)
	}
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
	// snapshot
	users := make([]*user.User, 0, len(r.Users))
	for _, u := range r.Users {
		users = append(users, u)
	}
	//房间存在先退出必须在锁内
	if joinuser.CurrentRoom != "" { //如果是带锁调用 发生死锁：同一个 goroutine 试图重复获取同一把锁（不可重入锁）
		s.leaveRoomUnsafe(joinuser) //同一把锁被同一个线程重复锁住
	}
	//join
	r.Users[joinuser.Name] = joinuser
	joinuser.CurrentRoom = roomName
	//复制要广播的人
	//1. map 不是线程安全的
	//2. 解锁以后其他 goroutine 可能修改 map
	//3. 遍历和修改同时发生会 panic
	//4. 所以先复制，再解锁

	s.mapLock.Unlock() //正常返回
	//广播消息
	msg := "[系统][" + joinuser.Name + "]" + "加入房间：" + roomName + "\n"
	s.RoomBroadcast(users, msg)
	joinuser.SendMsg("成功加入房间：" + roomName)
}

// 退出房间（外层）
func (s *Server) LeaveRoom(leaveuser *user.User) {
	s.mapLock.Lock()
	roomName := leaveuser.CurrentRoom
	r, ok := s.Rooms[roomName]
	if !ok {
		s.mapLock.Unlock()
		leaveuser.SendMsg("当前未加入房间")
		return
	}
	//make([]T, len, cap)
	users := make([]*user.User, 0, len(r.Users)) // snapshot
	for _, u := range r.Users {
		users = append(users, u)
	}
	//modify data
	s.leaveRoomUnsafe(leaveuser)
	s.mapLock.Unlock()
	msg := "[系统][" + leaveuser.Name + "]" + "离开房间：" + roomName + "\n"
	s.RoomBroadcast(users, msg)
	leaveuser.SendMsg("退出房间成功")
}

// 退出房间（内层）
func (s *Server) leaveRoomUnsafe(leaveuser *user.User) {
	roomName := leaveuser.CurrentRoom
	r := s.Rooms[roomName]
	delete(r.Users, leaveuser.Name)
	leaveuser.CurrentRoom = ""
	if len(r.Users) == 0 {
		delete(s.Rooms, roomName) //删除房间
	}
}

// 创建房间
func (s *Server) CreateRoom(createuser *user.User, roomName string) {
	s.mapLock.Lock()
	_, ok := s.Rooms[roomName] //检查重名
	if ok {
		s.mapLock.Unlock()
		createuser.SendMsg("房间名存在，请重试")
		return
	}
	//已进入房间 先退出
	oldRoom := createuser.CurrentRoom

	if oldRoom != "" {
		s.leaveRoomUnsafe(createuser)
	}

	r := room.NewRoom(roomName)           //返回一个房间对象
	s.Rooms[roomName] = r                 //房间放到Server中
	createuser.CurrentRoom = roomName     //记录用户当前房间
	r.Users[createuser.Name] = createuser //创建者进入房间
	s.mapLock.Unlock()
}

// 群聊功能
func (s *Server) RoomChat(sender *user.User, content string) {
	s.mapLock.Lock()

	if sender.CurrentRoom == "" {
		s.mapLock.Unlock()
		sender.SendMsg("当前未加入房间")
		return
	}
	r := s.Rooms[sender.CurrentRoom]
	// snapshot
	users := make([]*user.User, 0, len(r.Users))
	for _, u := range r.Users {
		users = append(users, u)
	}
	s.mapLock.Unlock()
	msg := "[" + sender.CurrentRoom + "][" + sender.Name + "] " + content
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

// 加入房间
func (s *Server) JoinRoom1(user, roomName string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	u, ok := s.OnlineUsers[user]
	if !ok {
		return "未找到用户", false
	}
	newRoom, ok := s.Rooms[roomName]
	if !ok {
		return "未找到房间", false
	}
	oldRoomName := u.CurrentRoom
	if oldRoomName == roomName {
		return "已加入房间", true
	}
	if oldRoomName != "" {
		oldRoom, _ := s.Rooms[oldRoomName]
		delete(oldRoom.Users, user) //删除旧房间
	}
	newRoom.Users[user] = u
	u.CurrentRoom = roomName
	return "加入房间", true
}

// 退出房间
func (s *Server) LeaveRoom1(user string) (string, bool) {
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	u, ok := s.OnlineUsers[user]
	if !ok {
		return "未找到用户", false
	}
	oldRoomName := u.CurrentRoom
	if oldRoomName == "" {
		return "未加入房间", true
	}
	if oldRoom, ok := s.Rooms[oldRoomName]; ok {
		delete(oldRoom.Users, user)
	}
	u.CurrentRoom = ""
	return "退出房间", true
}
