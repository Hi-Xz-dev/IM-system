package server

import "net"

// 在启动 goroutine 前登记，与 Shutdown 共用锁，避免等待开始后再加入连接。
func (s *Server) startTCPHandler(conn net.Conn) bool {
	s.lifecycleMu.Lock()
	if s.isShutdown {
		s.lifecycleMu.Unlock()
		_ = conn.Close()
		return false
	}
	s.tcpConnections[conn] = struct{}{}
	s.tcpHandlerWG.Add(1)
	s.lifecycleMu.Unlock()

	go func() {
		defer s.tcpHandlerWG.Done()
		defer func() {
			_ = conn.Close()
			s.lifecycleMu.Lock()
			delete(s.tcpConnections, conn)
			s.lifecycleMu.Unlock()
		}()
		s.Handler(conn)
	}()
	return true
}
