package server

import (
	"IM-system/user"
)

// 用户上线业务
func (s *Server) Online(user *user.User) {
	//用户上线，将用户加入OnlineUsers
	s.mapLock.Lock()
	s.OnlineUsers[user.Name] = user
	s.mapLock.Unlock()
	//广播当前用户上线信息
	s.BroadCast(user, "上线")
}

// 用户下线业务
func (s *Server) Offline(user *user.User) {
	//用户下线，将用户从OnlineUsers中删除
	s.mapLock.Lock()
	if user.CurrentRoom != "" {
		s.leaveRoomUnsafe(user)
	}
	delete(s.OnlineUsers, user.Name)
	s.mapLock.Unlock()
	s.BroadCast(user, "已下线")
}

// 当前位置
func (s *Server) Where(u *user.User) {
	s.mapLock.RLock()
	roomName := u.CurrentRoom
	s.mapLock.RUnlock()
	if roomName == "" {
		u.C <- "当前未加入房间\n"
		return
	}
	u.C <- "当前房间：" + roomName + "\n"
}

// 私聊功能
func (s *Server) PrivateChat(sender *user.User, targetName string, content string) {
	s.mapLock.RLock()
	targetUser, ok := s.OnlineUsers[targetName]
	s.mapLock.RUnlock()
	if !ok {
		sender.C <- "用户不存在，请重试\n"
		return
	}
	privateMsg := "[私聊][" + sender.Name + " -> " + targetName + "] " + content
	targetUser.C <- privateMsg
	sender.C <- "[系统] 私聊发送成功 -> " + targetName

}

// 用户改名业务
func (s *Server) Rename(usr *user.User, newName string) {
	s.mapLock.Lock()
	if _, ok := s.OnlineUsers[newName]; ok {
		s.mapLock.Unlock()
		usr.SendMsg("用户名已存在，请重试\n")
		return
	}
	oldName := usr.Name
	delete(s.OnlineUsers, oldName)
	usr.Name = newName
	s.OnlineUsers[newName] = usr
	//同步更新房间成员列表
	if usr.CurrentRoom != "" {
		if r, ok := s.Rooms[usr.CurrentRoom]; ok {
			delete(r.Users, oldName)
			r.Users[newName] = usr
		}
	}
	//snalshot
	users := make([]*user.User, 0, len(s.OnlineUsers))
	for _, u := range s.OnlineUsers {
		users = append(users, u)
	}

	s.mapLock.Unlock()

	// 4. IO（锁外）
	msg := "[系统] " + oldName + " 改名为 " + newName + "\n"
	s.RoomBroadcast(users, msg)
}

// Help
func (s *Server) Help(user *user.User) {
	user.C <- "======= 命令列表 =======\n"
	user.C <- "who                 查看在线用户\n"
	user.C <- "rename|名字         修改昵称\n"
	user.C <- "to|用户|消息        私聊\n"
	user.C <- "rooms               查看房间\n"
	user.C <- "create|房间         创建房间\n"
	user.C <- "join|房间           加入房间\n"
	user.C <- "leave               离开房间\n"
	user.C <- "room|消息           房间聊天\n"
	user.C <- "quit                退出系统\n"
	user.C <- "========================\n"
}

// 查找全部在线用户
func (s *Server) GetOnlineUsers() []string {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	users := make([]string, 0, len(s.OnlineUsers))
	for _, cli := range s.OnlineUsers {
		users = append(users, cli.Name)
	}

	return users
}
