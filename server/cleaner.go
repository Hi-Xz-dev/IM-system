package server

import(
	"IM-system/user"
	"time"
)
// 超时强踢功能
func (s *Server) CleanOnlineUser() {
	for {
		var kickList []*user.User
		s.mapLock.RLock()

		for _, u := range s.OnlineUsers {
			if u.ActiveTime != 0 &&
				time.Now().Unix()-u.ActiveTime > 6000 {
				kickList = append(kickList, u)
			}
		}
		s.mapLock.RUnlock()
		for _, u := range kickList {
			s.Offline(u)
		}
		time.Sleep(10 * time.Second)
	}
}