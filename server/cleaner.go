package server

import (
	"IM-system/user"
	"time"
)

// 超时强踢功能 当前实现（连接级踢出）
func (s *Server) CleanOnlineUser() {
	for {
		var kickList []*user.User
		s.mapLock.RLock()

		for _, clients := range s.OnlineUsers {
			for _, u := range clients {
				if u.ActiveTime != 0 &&
					time.Since(
						time.Unix(u.ActiveTime, 0),
					) > 10*time.Minute {
					kickList = append(
						kickList,
						u,
					)
				}
			}
		}
		s.mapLock.RUnlock()
		for _, u := range kickList {
			s.Offline(u)
		}
		time.Sleep(10 * time.Second)
	}
}
