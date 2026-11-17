package server

// 监听Disconnect 谁掉线就Offline()
func (s *Server) ListenDisconnect() {
	for {
		select {
		case <-s.done:
			return
		case usr, ok := <-s.Disconnect:
			if !ok {
				return
			}
			s.Offline(usr)
		}
	}
}
