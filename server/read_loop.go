package server

import (
	"IM-system/user"
	"bufio"
	"IM-system/internal/logger"
	"net"
	"strings"
)
//持续读取当前TCP客户端发来的信息，交给业务层处理
func (s *Server) readLoop(conn net.Conn, scanner *bufio.Scanner, usr *user.User, done chan<- struct{}) {
	defer close(done)

	for scanner.Scan() {//返回bool
		usr.UpdateActiveTime()
		//提取纯净信息
		msg := strings.TrimSpace(scanner.Text())
		if msg == "" {
			continue
		}
		if msg == "quit" {
			s.Offline(usr)
			return
		}
		s.DoMessage(usr, msg)
	}
	if err := scanner.Err(); err != nil {
	logger.Log.Error(
		"connection read failed",
		"user", usr.Nickname,
		"addr", usr.Addr,
		"error", err,
	)
}
	s.Offline(usr)
}
