package server

import (
	"IM-system/internal/logger"
	"IM-system/user"

	"context"
	"fmt"
	"strings"
	"time"
)

// 用户上线业务
func (s *Server) Online(usr *user.User) {
	s.mapLock.Lock()
	s.onlineLocked(usr, usr.Nickname)
	s.mapLock.Unlock()
	_ = s.BroadcastSystemMessage(usr, "上线")
}

// OnlineAuthenticated 使用当前在线资料，或在首端上线时从数据库加载。
func (s *Server) OnlineAuthenticated(ctx context.Context, usr *user.User) error {
	if err := s.registerAuthenticatedSession(ctx, usr); err != nil {
		return err
	}
	_ = s.BroadcastSystemMessage(usr, "上线")
	return nil
}

func (s *Server) registerAuthenticatedSession(ctx context.Context, usr *user.User) error {
	// 首端上线加载与改名串行，避免使用认证阶段读到的旧昵称。
	s.renameMu.Lock()
	defer s.renameMu.Unlock()

	s.mapLock.Lock()
	if profile := s.Profiles[usr.ID]; profile != nil {
		s.onlineLocked(usr, profile.Nickname)
		s.mapLock.Unlock()
		return nil
	}
	s.mapLock.Unlock()

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	nickname, err := s.authService.GetNickname(dbCtx, usr.ID)
	if err != nil {
		return fmt.Errorf("load online nickname: %w", err)
	}

	s.mapLock.Lock()
	s.onlineLocked(usr, nickname)
	s.mapLock.Unlock()
	return nil
}

func (s *Server) onlineLocked(usr *user.User, nickname string) {
	// 聊天室成员关系只在账号在线期间保留，新端继承其他在线端的房间。
	for _, session := range s.OnlineUsers[usr.ID] {
		for roomID := range session.JoinedRooms {
			if _, ok := s.Rooms[roomID]; ok {
				s.joinRoomUnsafe(usr, roomID)
			}
		}
	}
	s.OnlineUsers[usr.ID] =
		append(s.OnlineUsers[usr.ID], usr)

	if _, ok := s.Profiles[usr.ID]; !ok {
		s.Profiles[usr.ID] = &UserProfile{
			ID:       usr.ID,
			Nickname: nickname,
		}
	}
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

	roomIDs := make([]int64, 0, len(usr.JoinedRooms))
	for roomID := range usr.JoinedRooms {
		roomIDs = append(roomIDs, roomID)
	}
	// 断线仅清理当前 session，不触发同账号全端退房。
	for _, roomID := range roomIDs {
		s.removeRoomSessionUnsafe(usr, roomID)
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

	profile := s.Profiles[usr.ID]
	var nickname string
	if profile != nil {
		nickname = profile.Nickname
	}
	if len(users) == 0 {
		delete(s.OnlineUsers, usr.ID)
		delete(s.Profiles, usr.ID)
	} else {
		s.OnlineUsers[usr.ID] = users
	}

	s.mapLock.Unlock()

	usr.Close()

	if profile != nil {
		_ = s.broadcastSystemMessage(usr.ID, nickname, "已下线")
	}
}

// 查询用户加入房间列表
func (s *Server) Where(usr *user.User) {
	s.mapLock.RLock()
	rooms := make([]string, 0, len(usr.JoinedRooms))
	for roomID := range usr.JoinedRooms {
		r, ok := s.Rooms[roomID]
		if !ok {
			continue
		}
		rooms = append(rooms, r.Name)
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
	//对于锁外要遍历的切片，锁内需要创建独立的底层数组，才能让这份列表不被其他协程修改
	targetSessions = append([]*user.User(nil), targetSessions...)
	s.mapLock.RUnlock()
	if !ok || len(targetSessions) == 0 {

		err := s.SendSystemMessage(sender, "用户当前不在线")
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
		_ = s.SendSystemMessage(sender, "私聊发送失败，请稍后重试")
		return
	}

	if err := s.SendSystemMessage(
		sender, "私聊消息已进入发送队列",
	); err != nil {

		logger.Log.Warn(
			"send success message failed",
			"error",
			err,
		)
	}
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

// 获取Nickname
func (s *Server) GetNickname(userID int64) (string, bool) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()

	profile, ok := s.Profiles[userID]
	if !ok {
		return "", false
	}

	return profile.Nickname, true
}

// 用户历史信息
func (s *Server) PrivateHistory(usr *user.User, targetID int64) {
	page, err := s.messageService.GetPrivateHistory(
		context.Background(),
		usr.ID,
		targetID,
		nil,
	)

	if err != nil {
		logger.Log.Error(
			"get private history failed",
			"user_id", usr.ID,
			"target_id", targetID,
			"error", err,
		)

		_ = s.SendSystemMessage(
			usr,
			"获取历史消息失败",
		)
		return
	}

	for i := len(page.Messages) - 1; i >= 0; i-- {
		msg := page.Messages[i]

		text := fmt.Sprintf(
			"[%s] %s: %s",
			msg.Time.Format("15:04:05"),
			msg.FromNickname,
			msg.Content,
		)

		_ = s.SendSystemMessage(
			usr,
			text,
		)
	}
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
history-private|用户ID 用户历史信息
history-room|房间ID    房间历史信息
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

	return users
}

// 用户位置
func (s *Server) GetUserRoomsH(userID int64) ([]string, bool) {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	users, ok := s.OnlineUsers[userID]
	if !ok {
		return nil, false
	}

	roomSet := make(map[int64]struct{})
	for _, u := range users {
		for roomID := range u.JoinedRooms {
			roomSet[roomID] = struct{}{}
		}
	}
	rooms := make([]string, 0, len(roomSet))

	for roomID := range roomSet {

		r, ok := s.Rooms[roomID]
		if !ok {
			continue
		}

		rooms = append(rooms, r.Name)
	}
	return rooms, true
}
