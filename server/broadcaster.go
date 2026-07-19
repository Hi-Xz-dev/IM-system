package server

import (
	"IM-system/user"

)

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User 非阻塞广播
func (s *Server) ListenMessager() {
	for {
		msg := <-s.Message
		//将msg发送给全部的在线User
		s.mapLock.RLock()
		users := make([]*user.User, 0, len(s.OnlineUsers))
		for _, cli := range s.OnlineUsers {
			users = append(users, cli)
		}
		s.mapLock.RUnlock()
		for _, cli := range users {
			cli.SendMsg(msg)
		}
	}
}

// 广播消息方法
func (s *Server) BroadCast(user *user.User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Nickname + ":" + msg
	s.Message <- sendMsg

}

// 房间广播纯IO
func (s *Server) RoomBroadcast(users []*user.User, msg string) {
	for _, u := range users {
		u.SendMsg(msg)
	}
}
