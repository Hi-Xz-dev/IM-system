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
		return 0, "", errors.New("connection closed")
	}

	raw := scanner.Text()
	logger.Log.Info(
		"auth request",
		"raw", raw,
	)
	cmd := protocol.Parse(raw)
	logger.Log.Info(
		"parsed command",
		"type", cmd.Type,
		"args", cmd.Args,
	)
	if cmd.Type != domain.CmdAuth {
		return 0, "", errors.New("need auth first")
	}

	if len(cmd.Args) != 1 {
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
		logger.Log.Error(
			"authenticate failed",
			"error",
			err,
		)
		return 0, "", err
	}

	return userInfo.ID, userInfo.Nickname, nil
}
