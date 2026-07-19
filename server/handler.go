package server

import (
	"IM-system/user"
	"bufio"
	"net"
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
	//创建user
	usr := user.NewUser(
		conn,
		userID,
		nickname,
	)

	go usr.ListenMessage(s.Disconnect)
	//通知 Handler：读协程结束了，你也可以退出了 只做通知 内容无所谓 省内存
	done := make(chan struct{})
	//用户上线业务
	s.Online(usr)
	//认证成功
	conn.Write([]byte("[系统] 认证成功\n"))
	//启动读协程 负责读客户端发来的消息
	go s.readLoop(conn, scanner, usr, done)
	//等这个 Channel 被关闭。
	<-done
}
