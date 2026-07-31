package server

import (
	"IM-system/user"
	"bufio"
	"net"
	"IM-system/internal/connection"
)

// handler
func (s *Server) Handler(conn net.Conn) {

	scanner := bufio.NewScanner(conn)
	userID, nickname, err := s.authenticateConnection(scanner)

	if err != nil {
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

	go usr.ListenMessage(s.Disconnect)

	//用户上线业务
	s.Online(usr)
	//认证成功
	tcpConn.Write([]byte("[系统] 认证成功\n"))

	reader := connection.NewTCPReader(conn)
	//启动读协程 负责读客户端发来的消息
	s.ServerReader(reader,usr)

}
