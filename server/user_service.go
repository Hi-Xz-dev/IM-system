package server

import (
	"IM-system/user"
	"strings"
)

// 用户上线业务
func (s *Server) Online(user *user.User) {
	//用户上线，将用户加入OnlineUsers
	s.mapLock.Lock()
	s.OnlineUsers[user.Nickname] = user
	s.mapLock.Unlock()
	//广播当前用户上线信息
	s.BroadCast(user, "上线")
}

// 用户下线业务
func (s *Server) Offline(usr *user.User) {
	//用户下线，将用户从OnlineUsers中删除
	s.mapLock.Lock()
	if usr.IsClosed {
		s.mapLock.Unlock()
		return
	}
	usr.IsClosed = true
	//复制用户加入的全部房间
	roomNames := make([]string, 0, len(usr.JoinedRooms))
	for roomName := range usr.JoinedRooms {
		roomNames = append(roomNames, roomName)
	}
	//逐一退出
	for _, roomName := range roomNames {
		s.leaveRoomUnsafe(usr, roomName)
	}
	delete(s.OnlineUsers, usr.Nickname)
	s.mapLock.Unlock()
	usr.Close()
	s.BroadCast(usr, "已下线")
}

// 查询用户加入房间列表
func (s *Server) Where(usr *user.User) {
	s.mapLock.RLock()
	rooms := make([]string, 0, len(usr.JoinedRooms))
	for roomName := range usr.JoinedRooms{
		rooms = append(rooms, roomName)
	}
	s.mapLock.RUnlock()
	if len(rooms) == 0 {
		usr.SendMsg("当前未加入任何房间")
		return
	}
	usr.SendMsg("已加入房间：" + strings.Join(rooms,","))
}

// 私聊功能
func (s *Server) PrivateChat(sender *user.User, targetName string, content string) {
	s.mapLock.RLock()
	targetUser, ok := s.OnlineUsers[targetName]
	s.mapLock.RUnlock()
	if !ok {
		sender.SendMsg("用户不存在，请重试")
		return
	}
	privateMsg := "[私聊][" + sender.Nickname + " -> " + targetName + "] " + content
	targetUser.SendMsg(privateMsg)
	sender.SendMsg("[系统] 私聊发送成功 -> " + targetName)

}

// 用户改名业务
func (s *Server) Rename(usr *user.User, newName string) {
	s.mapLock.Lock()
	if _, ok := s.OnlineUsers[newName]; ok {
		s.mapLock.Unlock()
		usr.SendMsg("用户名已存在，请重试")
		return
	}
	oldName := usr.Nickname
	s.renameUserUnsafe(usr, oldName, newName)
	//snalshot
	users := make([]*user.User, 0, len(s.OnlineUsers))
	for _, u := range s.OnlineUsers {
		users = append(users, u)
	}
	s.mapLock.Unlock()
	msg := "[系统] " + oldName + " 改名为 " + newName + "\n"
	s.RoomBroadcast(users, msg)
}

// Help
func (s *Server) Help(user *user.User) {
	user.SendMsg(
`======= 命令列表 =======
who                   查看在线用户
rename|名字           修改昵称
to|用户|消息          私聊
rooms                 查看房间列表
create|房间           创建房间
join|房间             加入房间
leave|房间            退出房间
room|房间|消息        房间聊天
members|房间          房间成员
where                 查看已加入的房间
help                  命令列表
quit                  退出系统
========================`)
}

// 查找全部在线用户
func (s *Server) GetOnlineUsers() []string {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	users := make([]string, 0, len(s.OnlineUsers))
	for _, cli := range s.OnlineUsers {
		users = append(users, cli.Nickname)
	}

	return users
}
//===============Unsafe================
//同步更新房间成员列表
func (s *Server) renameUserRoomsUnsafe(usr *user.User, oldName, newName string){
	for roomName := range usr.JoinedRooms{
		r, ok := s.Rooms[roomName]
		if !ok{
			continue
		}
		delete(r.Users, oldName)
		r.Users[newName] = usr
	}
}
func (s *Server) renameUserUnsafe(usr *user.User, oldName, newName string){
	delete(s.OnlineUsers, oldName)
	usr.Nickname = newName
	s.OnlineUsers[newName] = usr
	s.renameUserRoomsUnsafe(usr, oldName, newName)
	
}

//===============HTTP==================
//用户改名
func (s *Server) RenameByName(oldName, newName string)(string, bool){
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	usr, ok := s.OnlineUsers[oldName]
	if !ok{
		return "未找到用户", false
	}
	if _, ok := s.OnlineUsers[newName];ok{
		return "用户名已存在", false
	}
	s.renameUserUnsafe(usr, oldName, newName)
	
	return "修改用户名成功", true
}
//用户位置
func (s *Server) GetUserRooms(userName string) ([]string, bool){
	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	u, ok := s.OnlineUsers[userName]
	if !ok{
		return nil, false
	}
	rooms := make([]string, 0, len(u.JoinedRooms))
	for roomName := range u.JoinedRooms{
		rooms = append(rooms, roomName)
	}
	if len(rooms) == 0 {
		return []string{}, true
	}
	return rooms, true	
}

