package server

import (
	"IM-system/internal/logger"
	"IM-system/user"
	"fmt"
	"net"
)

// 启动服务器接口
func (s *Server) Start() {
	//在指定 IP 和端口创建 TCP 监听器 listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		logger.Log.Error("net.Listen err","error", err)
		return
	}
	s.listener = listener
	logger.Log.Info(
	"tcp server started",
	"ip", s.IP,
	"port", s.Port,
	)
	//close listen socket
	defer listener.Close()
	//定时清理超时用户
	go s.CleanOnlineUser()
	//等待断线用户事件，然后执行 Offline
	go s.ListenDisconnect()
	//等待广播消息，然后投递给在线用户
	go s.ListenMessager()
	for {
		//accept等待 TCP三次握手
		conn, err := listener.Accept()
		if err != nil {
			if s.IsShutdown {
				logger.Log.Info("server shutdown, stop accepting connections",)
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
