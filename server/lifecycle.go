package server

import (
	"IM-system/internal/logger"
	"IM-system/user"
	"fmt"
	"net"
)

// 启动服务器接口
func (s *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	s.listener = listener
	//close listen socket
	defer listener.Close()
	//启动超时踢人功能
	go s.CleanOnlineUser()
	go s.ListenDisconnect()
	//启动监听Message的goroutine
	go s.ListenMessager()
	for {
		//accept等待 TCP三次握手
		conn, err := listener.Accept()
		if err != nil {
			if s.IsShutdown {
				fmt.Println("server shutdown, stop accepting connections")
				return
			}
			logger.Log.Error("accept failed", "error", err)
			continue
		}
		//do handler
		go s.Handler(conn) //goroutine
	}
}

func (s *Server) Shutdown() {
	s.IsShutdown = true
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.mapLock.RLock()
	users := make([]*user.User, 0, len(s.OnlineUsers))
	for _, u := range s.OnlineUsers {
		users = append(users, u)
	}
	s.mapLock.RUnlock()
	for _, u := range users {
		u.SendMsg("[系统] 服务器正在关闭")
		s.Offline(u)
	}
}
