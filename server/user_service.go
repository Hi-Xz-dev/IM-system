package server

import (
	"IM-system/internal/logger"
	"IM-system/user"

	"strings"
)

// 用户上线业务
func (s *Server) Online(usr *user.User) {
	//用户上线，将用户加入OnlineUsers
	s.mapLock.Lock()
	s.OnlineUsers[usr.ID] =
		append(s.OnlineUsers[usr.ID], usr)
	s.mapLock.Unlock()
	//广播当前用户上线信息
	s.BroadcastSystemMessage(usr, "上线")
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

	roomNames := make([]string, 0, len(usr.JoinedRooms))
	for roomName := range usr.JoinedRooms {
		roomNames = append(roomNames, roomName)
	}
	//逐一退出
	for _, roomName := range roomNames {
		s.leaveRoomUnsafe(usr, roomName)
	}
	//删除当前链接
	users := s.OnlineUsers[usr.ID]

	for i, u := range users {
		if u == usr {
			//删除元素
			users = append(users[:i], users[i+1:]...)
			break
		}
	}

	if len(users) == 0 {
		delete(s.OnlineUsers, usr.ID)
	} else {
		s.OnlineUsers[usr.ID] = users
	}

	s.mapLock.Unlock()

	usr.Close()

	s.BroadcastSystemMessage(usr, "已下线")
}

// 查询用户加入房间列表
func (s *Server) Where(usr *user.User) {
	s.mapLock.RLock()
	rooms := make([]string, 0, len(usr.JoinedRooms))
	for roomName := range usr.JoinedRooms {
		rooms = append(rooms, roomName)
	}
	s.mapLock.RUnlock()
	if len(rooms) == 0 {
		_ = s.SendSystemMessage(usr, "当前未加入任何房间")
		return
	}
	_ = s.SendSystemMessage(usr, "已加入房间："+strings.Join(rooms, ","))
}

// 私聊功能
func (s *Server) PrivateChat(sender *user.User, targetID int64, content string) {
	s.mapLock.RLock()
	targetSessions, ok := s.OnlineUsers[targetID]
	s.mapLock.RUnlock()
	if !ok || len(targetSessions) == 0 {

		err := s.SendSystemMessage(sender, "用户不存在，请重试")
		if err != nil {
			logger.Log.Error(
				"send system message failed",
				"error",
				err,
			)
		}
		return
	}
	if err := s.SendPrivateMessage(
		sender, targetSessions, targetID, content,
	); err != nil {
		logger.Log.Error(
			"send private message failed",
			"error",
			err,
		)
	}

	if err := s.SendSystemMessage(
		sender, "私聊发送成功",
	); err != nil {

		logger.Log.Warn(
			"send success message failed",
			"error",
			err,
		)
	}
}

// 用户改名业务
func (s *Server) Rename(usr *user.User, newName string) {
	s.mapLock.Lock()
	if s.nameExistsUnsafe(newName) {
		s.mapLock.Unlock()

		_ = s.SendSystemMessage(
			usr,
			"用户名已存在，请重试",
		)
		return
	}

	oldName := usr.Nickname

	users, ok := s.OnlineUsers[usr.ID]

	if !ok {
		s.mapLock.Unlock()
		return
	}

	for _, u := range users {
		u.Nickname = newName
	}

	s.mapLock.Unlock()

	if err := s.BroadcastSystemMessage(
		usr,
		oldName+" 改名为 "+newName,
	); err != nil {

		logger.Log.Error(
			"broadcast rename failed",
			"error",
			err,
		)
	}
}

// 检查重名
func (s *Server) nameExistsUnsafe(name string) bool {

	for _, users := range s.OnlineUsers {

		for _, usr := range users {

			if usr.Nickname == name {
				return true
			}
		}
	}

	return false
}

// 返回在线用户切片
func (s *Server) getOnlineSessionsUnsafe() []*user.User {

	users := make([]*user.User, 0)

	for _, clients := range s.OnlineUsers {
		for _, cli := range clients {
			users = append(users, cli)
		}
	}

	return users
}

// Help
func (s *Server) Help(user *user.User) {
	_ = s.SendSystemMessage(user,
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

// ===============HTTP==================

// 查找全部在线用户
func (s *Server) GetOnlineUsers() []OnlineUser {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	seen := make(map[int64]string)
	for id, clients := range s.OnlineUsers {
		for _, cli := range clients {
			seen[id] = cli.Nickname
		}
	}
	users := make([]OnlineUser, 0, len(seen))
	for id, nick := range seen {
		users = append(users, OnlineUser{ID: id, Nickname: nick})
	}
	return users
}

// 用户改名
func (s *Server) RenameH(userID int64, newName string) (string, bool) {
	s.mapLock.Lock()
	users, ok := s.OnlineUsers[userID]
	if !ok {
		s.mapLock.Unlock()
		return "未找到用户", false
	}
	for _, us := range s.OnlineUsers {
		for _, u := range us {
			if u.Nickname == newName {
				s.mapLock.Unlock()
				return "用户名已存在", false
			}
		}
	}
	oldName := users[0].Nickname
	for _, u := range users {
		u.Nickname = newName
	}
	s.mapLock.Unlock()
	// 广播改名通知
	_ = s.BroadcastSystemMessage(users[0], oldName+" 改名为 "+newName)
	return "修改用户名成功", true
}

// 用户位置
func (s *Server) GetUserRoomsH(userID int64) ([]string, bool) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	users, ok := s.OnlineUsers[userID]
	if !ok {
		return nil, false
	}

	roomSet := make(map[string]struct{})
	for _, u := range users {
		for roomName := range u.JoinedRooms {
			roomSet[roomName] = struct{}{}
		}
	}
	rooms := make([]string, 0, len(roomSet))
	for name := range roomSet {
		rooms = append(rooms, name)
	}
	return rooms, true
}
