package server

import (
	"IM-system/internal/logger"
	"IM-system/user"
	"fmt"
	"net"
)

// 启动服务器接口
func (s *Server) Start() error {
	//在指定 IP 和端口创建 TCP 监听器 listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}
	if !s.registerListener(listener) {
		_ = listener.Close()
		return nil
	}
	logger.Log.Info(
		"tcp server started",
		"ip", s.IP,
		"port", s.Port,
	)
	//close listen socket
	defer listener.Close()
	s.startBackgroundWorkers()
	for {
		//accept等待 TCP三次握手
		conn, err := listener.Accept()
		if err != nil {
			if s.isShuttingDown() {
				logger.Log.Info("server shutdown, stop accepting connections")
				return nil
			}
			return fmt.Errorf("accept tcp: %w", err)
		}
		//do handler
		go s.Handler(conn) //goroutine
	}
}

// 关闭全部链接
func (s *Server) Shutdown() {
	s.lifecycleMu.Lock()
	if !s.isShutdown {
		s.isShutdown = true
		if s.done != nil {
			close(s.done)
		}
	}
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
	s.workerWG.Wait()
}

// 与关闭状态共用锁，确保 Shutdown 开始等待后，不会再增加后台任务。
func (s *Server) startBackgroundWorkers() {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.isShutdown || s.workersStarted {
		return
	}
	s.workersStarted = true
	for _, work := range []func(){s.CleanOnlineUser, s.ListenDisconnect, s.ListenMessager} {
		s.workerWG.Add(1)
		go func(work func()) {
			defer s.workerWG.Done()
			work()
		}(work)
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
