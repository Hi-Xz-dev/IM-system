package server

import (
	"IM-system/internal/connection"
	"IM-system/user"
	"context"
	"net"
	"time"
)

// handler
func (s *Server) Handler(conn net.Conn) {
	defer conn.Close()
	tcpConn := connection.NewTCPConnection(conn)
	if err := conn.SetReadDeadline(time.Now().Add(connection.AuthReadTimeout)); err != nil {
		return
	}

	reader := connection.NewTCPReader(conn)
	userID, nickname, err := s.authenticateConnection(reader)

	if err != nil {
		if s.isShuttingDown() {
			return
		}
		_ = tcpConn.Write([]byte("[系统] 认证失败\n"))
		return
	}
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return
	}
	addr := conn.RemoteAddr().String()
	//创建user
	usr := user.NewUser(
		tcpConn,
		userID,
		nickname,
		addr,
	)
	if s.isShuttingDown() {
		usr.Close()
		return
	}
	if err := s.OnlineAuthenticated(context.Background(), usr); err != nil {
		_ = tcpConn.Write([]byte("[系统] 上线失败，请稍后重试\n"))
		return
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		usr.ListenMessage(s.Disconnect)
	}()
	defer func() {
		s.Offline(usr)
		usr.Close()
		<-writerDone
	}()

	//认证成功
	tcpConn.Write([]byte("[系统] 认证成功\n"))

	//启动读协程 负责读客户端发来的消息
	s.ServerReader(reader, usr)

}
