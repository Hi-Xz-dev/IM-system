package server


// 监听Disconnect 谁掉线就Offline()
func (s *Server) ListenDisconnect() {
	for usr := range s.Disconnect {
		s.Offline(usr)
	}
}