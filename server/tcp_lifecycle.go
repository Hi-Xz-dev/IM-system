package server

import "net"

// 在启动 goroutine 前登记，与 Shutdown 共用锁，避免等待开始后再加入连接。
func (s *Server) startTCPHandler(conn net.Conn) bool {
	finish, ok := s.TrackConnection(conn)
	if !ok {
		_ = conn.Close()
		return false
	}
	go func() {
		defer finish()
		s.Handler(conn)
	}()
	return true
}
