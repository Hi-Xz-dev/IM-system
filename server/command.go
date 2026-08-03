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
		s.handlerLeaveRoom(usr, cmd.Args)
	//help
	case domain.CmdHelp:
		s.Help(usr)
	//当前位置
	case domain.CmdWhere:
		s.Where(usr)
	//房间人数
	case domain.CmdMembers:
		s.handlerMembers(usr, cmd.Args)
	default:
		s.BroadcastSystemMessage(usr, msg)
	}
}

// 拆handler-XX函数
func (s *Server) handlerWho(usr *user.User) {
	s.mapLock.RLock()
	users := make(map[int64]*user.User)

	for _, clients := range s.OnlineUsers {

		for _, cli := range clients {

			users[cli.ID] = cli
		}
	}

	s.mapLock.RUnlock()

	for _, cli := range users {
		_ = s.SendSystemMessage(usr, cli.Nickname+" 在线")
	}

}
func (s *Server) handlerRename(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: rename|新名字")
		return
	}

	s.Rename(usr, args[0])
}

func (s *Server) handlerPrivateChat(usr *user.User, args []string) {
	if len(args) != 2 {
		_ = s.SendSystemMessage(usr, "用法: to|用户ID|消息")
		return
	}
	userID, err := protocol.ParseUserID(args[0])
	if err != nil {
		_ = s.SendSystemMessage(usr, "用户ID错误")
		return
	}
	s.PrivateChat(usr, userID, args[1])
}

func (s *Server) handlerCreate(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: create|房间名")
		return
	}

	s.CreateRoom(usr, args[0])
}
func (s *Server) handlerJoin(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: join|房间名")
		return
	}

	s.JoinRoom(usr, args[0])
}

// 房间聊天
func (s *Server) handlerRoomchat(usr *user.User, args []string) {
	if len(args) != 2 {
		_ = s.SendSystemMessage(usr, "用法: room|房间名|消息内容")
		return
	}

	s.RoomChat(usr, args[0], args[1])
}

// 离开房间
func (s *Server) handlerLeaveRoom(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: leave|房间名")
		return
	}

	s.LeaveRoom(usr, args[0])
}

// 房间成员
func (s *Server) handlerMembers(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: members|房间名")
		return
	}
	s.Members(usr, args[0])
}
