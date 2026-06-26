package server

import (
	"IM-system/user"
	"strings"

)

// 用户处理信息业务自定义IM协议 协议设计
func (s *Server) DoMessage(usr *user.User, msg string) {
	switch {
	//用户查询当前在线人数
	case msg == "who":
		s.handlerWho(usr)
	//更改用户名
	case strings.HasPrefix(msg, "rename|"):
		s.handlerRename(usr, msg)
	//私聊
	case strings.HasPrefix(msg, "to|"):
		s.handlerPrivateChat(usr, msg)
	//显示房间
	case msg == "rooms":
		s.ShowRooms(usr)
	//创建房间
	case strings.HasPrefix(msg, "create|"):
		s.handlerCreate(usr, msg)
	//加入房间
	case strings.HasPrefix(msg, "join|"):
		s.handlerJoin(usr, msg)
	//群聊功能
	case strings.HasPrefix(msg, "room|"):
		s.handlerRoomchat(usr, msg)
	//退出房间
	case msg == "leave":
		s.LeaveRoom(usr)
	//help
	case msg == "help":
		s.Help(usr)
	//当前位置
	case msg == "where":
		s.Where(usr)
	//房间人数
	case msg == "members":
		s.Members(usr)
	default:
		s.BroadCast(usr, msg)
	}
}

//拆handler-XX函数
func (s *Server) handlerWho(usr *user.User) {
		s.mapLock.RLock()
		users := make([]*user.User, 0, len(s.OnlineUsers))
		for _, cli := range s.OnlineUsers {
			users = append(users, cli)
		}

		s.mapLock.RUnlock()
		for _, cli := range users {
			onlineMsg := "[" + cli.Addr + "]" + cli.Name + ":" + "在线\n"
			usr.SendMsg(onlineMsg)
		}
}
func (s *Server) handlerRename(usr *user.User, msg string){
		parts := strings.Split(msg, "|")
		if len(parts) == 2 {
			s.Rename(usr, parts[1])
		} else {
			usr.SendMsg(("格式错误:rename|name\n"))
		}
}
func (s *Server) handlerPrivateChat(usr *user.User, msg string){
		parts := strings.Split(msg, "|")
		if len(parts) == 3 {
			s.PrivateChat(usr, parts[1], parts[2])
		} else {
			usr.SendMsg(("格式错误:to|name|msg\n"))
		}
}
func (s *Server) handlerCreate(usr *user.User, msg string){
		parts := strings.Split(msg, "|")
		if len(parts) == 2 && parts[1] != "" { //防止空房间名
			s.CreateRoom(usr, parts[1])
		} else {
			usr.C <- "格式错误:create|room\n"
		}
}
func (s *Server) handlerJoin(usr *user.User, msg string){
		parts := strings.Split(msg, "|")
		if len(parts) == 2 && parts[1] != "" { //防止空房间名
			s.JoinRoom(usr, parts[1])
		} else {
			usr.C <- "格式错误:join|room\n"
		}
}
func (s *Server) handlerRoomchat(usr *user.User, msg string){
		parts := strings.Split(msg, "|")
		if len(parts) == 2 && parts[1] != "" { //防止空房间名
			s.RoomChat(usr, parts[1])
		} else {
			usr.C <- "格式错误:room|room\n"
		}
}