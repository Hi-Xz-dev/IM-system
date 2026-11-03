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
		logger.Log.Error("net.Listen err", "error", err)
		return
	}
	if !s.registerListener(listener) {
		_ = listener.Close()
		return
	}
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
	//异步信息持久化
	go s.ListenPersistMessage()
	for {
		//accept等待 TCP三次握手
		conn, err := listener.Accept()
		if err != nil {
			if s.isShuttingDown() {
				logger.Log.Info("server shutdown, stop accepting connections")
				return
			}
			logger.Log.Error("accept failed", "error", err)
			continue
		}
		//do handler
		go s.Handler(conn) //goroutine
	}
}

// 关闭全部链接
func (s *Server) Shutdown() {
	s.lifecycleMu.Lock()
	s.isShutdown = true
	listener := s.listener
	s.lifecycleMu.Unlock()
	if listener != nil {
		_ = listener.Close()
	}
	s.mapLock.RLock()

	users := make([]*user.User, 0)
	for _, clients := range s.OnlineUsers {

		for _, usr := range clients {

			users = append(users, usr)

		}
	}
	s.mapLock.RUnlock()
	for _, u := range users {
		_ = s.SendSystemMessage(u, "服务器正在关闭")
		s.Offline(u)
	}
}

// 发布监听器与关闭状态使用同一把锁，防止 Shutdown 漏掉刚创建的监听器。
func (s *Server) registerListener(listener net.Listener) bool {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.isShutdown {
		return false
	}
	s.listener = listener
	return true
}

func (s *Server) isShuttingDown() bool {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	return s.isShutdown
}
