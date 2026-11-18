package server

import (
	"IM-system/internal/connection"
	"IM-system/user"
	"bufio"
	"net"
)

// handler
func (s *Server) Handler(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	userID, nickname, err := s.authenticateConnection(scanner)

	if err != nil {
		if s.isShuttingDown() {
			return
		}
		conn.Write([]byte("[系统] 认证失败\n"))
		conn.Close()
		return
	}
	addr := conn.RemoteAddr().String()
	tcpConn := connection.NewTCPConnection(conn)
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

	//用户上线业务
	s.Online(usr)
	//认证成功
	tcpConn.Write([]byte("[系统] 认证成功\n"))

	reader := connection.NewTCPReader(conn)
	//启动读协程 负责读客户端发来的消息
	s.ServerReader(reader, usr)

}
