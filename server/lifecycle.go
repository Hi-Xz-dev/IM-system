package server

import (
	"IM-system/internal/logger"
	"IM-system/user"
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// 启动服务器接口
func (s *Server) Start() error {
	// 监听器为接入的 TCP 连接设置系统保活探测。
	listenConfig := net.ListenConfig{
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   true, //启用 TCP 保活探测
			Idle:     30 * time.Second, //连接空闲 30 秒后，系统发出第一次探测
			Interval: 10 * time.Second, //探测没回应时，按 10 秒间隔重试，连续 3 次没回应后由系统判定连接失效。
			Count:    3, //没收到回应的探测包
		},
	}
	listener, err := listenConfig.Listen(context.Background(), "tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
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
		s.startTCPHandler(conn)
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
	connections := make([]io.Closer, 0, len(s.connections))
	for conn := range s.connections {
		connections = append(connections, conn)
	}
	s.lifecycleMu.Unlock()
	if listener != nil {
		_ = listener.Close()
	}
	// 包括尚未认证的连接；关闭 socket 可解除认证/读循环的网络等待。
	for _, conn := range connections {
		_ = conn.Close()
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
	s.connectionWG.Wait()
}

// TrackConnection 登记一条连接。调用方负责在所有读写任务结束后调用返回的清理函数。
func (s *Server) TrackConnection(conn io.Closer) (finish func(), ok bool) {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.isShutdown {
		return nil, false
	}
	s.connections[conn] = struct{}{}
	s.connectionWG.Add(1)
	return func() {
		_ = conn.Close()
		s.lifecycleMu.Lock()
		delete(s.connections, conn)
		s.lifecycleMu.Unlock()
		s.connectionWG.Done()
	}, true
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
