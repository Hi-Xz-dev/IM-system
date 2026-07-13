package server

import(
	"IM-system/user"
	"net"
)
// handler
func (s *Server) Handler(conn net.Conn) {
	//把底层 TCP 连接包装成业务用户对象
	usr := user.NewUser(conn)
	go usr.ListenMessage(s.Disconnect)
	//通知 Handler：读协程结束了，你也可以退出了 只做通知 内容无所谓 省内存
	done := make(chan struct{})
	//用户上线业务
	s.Online(usr)
	//启动读协程 负责读客户端发来的消息
	go s.readLoop(conn, usr, done)
	//等这个 Channel 被关闭。
	<-done 
}
