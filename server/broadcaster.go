package server

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部的在线User 非阻塞广播
func (s *Server) ListenMessager() {
	for {
		var msg string
		select {
		case <-s.done:
			return
		case next, ok := <-s.Message:
			if !ok {
				return
			}
			msg = next
		}
		//将msg发送给全部的在线User
		s.mapLock.RLock()
		users := s.getOnlineSessionsUnsafe()

		s.mapLock.RUnlock()
		for _, cli := range users {
			cli.SendMsg(msg)
		}
	}
}
