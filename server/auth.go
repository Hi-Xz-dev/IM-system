package server

import (
	"IM-system/internal/domain"
	"IM-system/internal/logger"
	"IM-system/internal/protocol"
	"bufio"
	"context"
	"errors"
)

func (s *Server) authenticateConnection(
	scanner *bufio.Scanner,
) (int64, string, error) {
	//读取第一条信息
	if !scanner.Scan() {
		logger.Log.Warn("authentication failed", "protocol", "tcp", "reason", "read_failed")
		return 0, "", errors.New("connection closed")
	}

	raw := scanner.Text()
	cmd := protocol.Parse(raw)
	if cmd.Type != domain.CmdAuth {
		logger.Log.Warn("authentication failed", "protocol", "tcp", "reason", "auth_required")
		return 0, "", errors.New("need auth first")
	}

	if len(cmd.Args) != 1 {
		logger.Log.Warn("authentication failed", "protocol", "tcp", "reason", "invalid_format")
		return 0, "", errors.New("invalid auth format")
	}

	token := cmd.Args[0]

	//调用认证服务
	userInfo, err :=
		s.authService.Authenticate(
			context.Background(),
			token,
		)

	if err != nil {
		logger.Log.Warn("authentication failed", "protocol", "tcp", "reason", "authentication_rejected")
		return 0, "", err
	}

	logger.Log.Info("authentication succeeded", "protocol", "tcp", "user_id", userInfo.ID)
	return userInfo.ID, userInfo.Nickname, nil
}
