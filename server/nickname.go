package server

import (
	"context"
	"errors"
	"time"

	"IM-system/internal/auth"
	"IM-system/internal/logger"
	"IM-system/user"
)

var ErrRenameUserOffline = errors.New("user is offline")

func (s *Server) Rename(usr *user.User, newName string) {
	err := s.RenameH(context.Background(), usr.ID, newName)
	switch {
	case err == nil:
	case errors.Is(err, auth.ErrInvalidNickname):
		_ = s.SendSystemMessage(usr, "昵称须为 1 到 64 个字符")
	case errors.Is(err, ErrRenameUserOffline):
		_ = s.SendSystemMessage(usr, "未找到在线用户")
	default:
		logger.Log.Error("rename user failed", "user_id", usr.ID, "error", err)
		_ = s.SendSystemMessage(usr, "修改昵称失败，请稍后重试")
	}
}

func (s *Server) RenameH(ctx context.Context, userID int64, newName string) error {
	session, oldName, nickname, err := s.updateNickname(ctx, userID, newName)
	if err != nil {
		return err
	}
	if session != nil && oldName != nickname {
		_ = s.BroadcastSystemMessage(session, oldName+" 改名为 "+nickname)
	}
	return nil
}

func (s *Server) updateNickname(ctx context.Context, userID int64, newName string) (*user.User, string, string, error) {
	// 串行处理改名，避免两次更新以不同顺序写入数据库和内存。
	s.renameMu.Lock()
	defer s.renameMu.Unlock()

	s.mapLock.RLock()
	profile := s.Profiles[userID]
	if len(s.OnlineUsers[userID]) == 0 || profile == nil {
		s.mapLock.RUnlock()
		return nil, "", "", ErrRenameUserOffline
	}
	oldName := profile.Nickname
	s.mapLock.RUnlock()

	if s.authService == nil {
		return nil, "", "", errors.New("user account service is not configured")
	}
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	nickname, err := s.authService.UpdateNickname(dbCtx, userID, newName)
	if err != nil {
		return nil, "", "", err
	}

	s.mapLock.Lock()
	defer s.mapLock.Unlock()
	profile.Nickname = nickname
	var session *user.User
	if sessions := s.OnlineUsers[userID]; len(sessions) > 0 {
		session = sessions[0]
	}
	return session, oldName, nickname, nil
}
