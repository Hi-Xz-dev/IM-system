package server

import (
	"IM-system/internal/domain"
	"IM-system/internal/protocol"
	"IM-system/user"

)

// 用户处理信息业务自定义IM协议 协议设计
func (s *Server) DoMessage(usr *user.User, msg string) {
	cmd := protocol.Parse(msg)
	switch cmd.Type {
	//用户查询当前在线人数
	case domain.CmdWho:
		s.handlerWho(usr)
	//更改用户名
	case domain.CmdRename:
		s.handlerRename(usr, cmd.Args)
	//私聊
	case domain.CmdPrivate:
		s.handlerPrivateChat(usr, cmd.Args)
	//显示房间
	case domain.CmdRooms:
		s.ShowRooms(usr)
	//创建房间
	case domain.CmdCreate:
		s.handlerCreate(usr, cmd.Args)
	//加入房间
	case domain.CmdJoin:
		s.handlerJoin(usr, cmd.Args)
	//群聊功能
	case domain.CmdRoom:
		s.handlerRoomchat(usr, cmd.Args)
	//退出房间
	case domain.CmdLeave:
		s.LeaveRoom(usr)
	//help
	case domain.CmdHelp:
		s.Help(usr)
	//当前位置
	case domain.CmdWhere:
		s.Where(usr)
	//房间人数
	case domain.CmdMembers:
		s.Members(usr)
	default:
		s.BroadCast(usr, msg)
	}
}

// 拆handler-XX函数
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
func (s *Server) handlerRename(usr *user.User, args []string) {
	if len(args) != 1 {
		usr.SendMsg("[系统] 用法: rename|新名字\n")
		return
	}

	s.Rename(usr, args[0])
}
func (s *Server) handlerPrivateChat(usr *user.User, args []string) {
	if len(args) != 2 {
		usr.SendMsg("[系统] 用法: to|用户名|消息\n")
		return
	}

	s.PrivateChat(usr, args[0], args[1])
}

func (s *Server) handlerCreate(usr *user.User, args []string) {
	if len(args) != 1 {
		usr.SendMsg("[系统] 用法: create|房间名\n")
		return
	}

	s.CreateRoom(usr, args[0])
}
func (s *Server) handlerJoin(usr *user.User, args []string) {
	if len(args) != 1 {
		usr.SendMsg("[系统] 用法: join|房间名\n")
		return
	}

	s.JoinRoom(usr, args[0])
}
func (s *Server) handlerRoomchat(usr *user.User, args []string) {
	if len(args) != 1 {
		usr.SendMsg("[系统] 用法: room|消息内容\n")
		return
	}

	s.RoomChat(usr, args[0])
}
