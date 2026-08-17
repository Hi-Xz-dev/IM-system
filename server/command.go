package server

import (
	"IM-system/internal/domain"
	"IM-system/internal/logger"
	"IM-system/internal/protocol"
	"IM-system/internal/repository"
	"IM-system/user"

	"errors"
	"strconv"
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

	users := make([]OnlineUser, 0, len(s.OnlineUsers))

	for userID := range s.OnlineUsers {
		profile, ok := s.Profiles[userID]
		if !ok {
			continue
		}

		users = append(users, OnlineUser{
			ID:       userID,
			Nickname: profile.Nickname,
		})
	}

	s.mapLock.RUnlock()

	for _, u := range users {
		_ = s.SendSystemMessage(
			usr,
			u.Nickname+" 在线",
		)
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
		s.SendSystemMessage(usr, "用法: create|房间名")
		return
	}

	if err := s.CreateRoom(usr, args[0]); err != nil {

		if errors.Is(err, repository.ErrRoomAlreadyExists) {
			s.SendSystemMessage(
				usr,
				"房间名已存在",
			)
			return
		}

		logger.Log.Error(
			"create room failed",
			"user_id", usr.ID,
			"room_name", args[0],
			"error", err,
		)
		s.SendSystemMessage(
			usr,
			"创建房间失败，请稍后重试",
		)
		return
	}

	s.SendSystemMessage(
		usr,
		"创建房间成功",
	)
}
func (s *Server) handlerJoin(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: join|房间ID")
		return
	}

	roomID, err := strconv.ParseInt(args[0], 10, 64)

	if err != nil {
		_ = s.SendSystemMessage(usr, "房间ID格式错误")
		return
	}

	s.JoinRoom(usr, roomID)
}

// 房间聊天
func (s *Server) handlerRoomchat(usr *user.User, args []string) {
	if len(args) != 2 {
		_ = s.SendSystemMessage(usr, "用法: room|房间ID|消息内容")
		return
	}

	roomID, err := strconv.ParseInt(args[0], 10, 64)

	if err != nil {
		_ = s.SendSystemMessage(usr, "房间ID格式错误")
		return
	}
	s.RoomChat(usr, roomID, args[1])
}

// 离开房间
func (s *Server) handlerLeaveRoom(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: leave|房间ID")
		return
	}

	roomID, err := strconv.ParseInt(args[0], 10, 64)

	if err != nil {
		_ = s.SendSystemMessage(usr, "房间ID格式错误")
		return
	}

	s.LeaveRoom(usr, roomID)
}

// 房间成员
func (s *Server) handlerMembers(usr *user.User, args []string) {
	if len(args) != 1 {
		_ = s.SendSystemMessage(usr, "用法: members|房间ID")
		return
	}
	roomID, err := strconv.ParseInt(args[0], 10, 64)

	if err != nil {
		_ = s.SendSystemMessage(usr, "房间ID格式错误")
		return
	}

	s.Members(usr, roomID)
}
