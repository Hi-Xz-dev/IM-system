package server

import (
	"IM-system/user"
	"time"
)

// 超时强踢功能 当前实现（连接级踢出）
func (s *Server) CleanOnlineUser() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		default:
		}
		kickList := s.inactiveUsers(time.Now())
		for _, u := range kickList {
			s.Offline(u)
		}
		select {
		case <-s.done:
			return
		case <-ticker.C:
		}
	}
}

func (s *Server) inactiveUsers(now time.Time) []*user.User {
	s.mapLock.RLock()
	defer s.mapLock.RUnlock()
	var users []*user.User
	for _, sessions := range s.OnlineUsers {
		for _, u := range sessions {
			active := u.LastActiveTime()
			if active != 0 && now.Sub(time.Unix(active, 0)) > 10*time.Minute {
				users = append(users, u)
			}
		}
	}
	return users
}
