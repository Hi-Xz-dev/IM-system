package server

import (
	"IM-system/internal/connection"
	"IM-system/user"

	"strings"
)

func (s *Server) ServerReader(reader connection.Reader, usr *user.User) {
	for {
		msg, err := reader.Read()

		if err != nil {
			s.Offline(usr)
			return
		}

		usr.UpdateActiveTime()
		//提取纯净信息
		msg = strings.TrimSpace(msg)

		if msg == "" {
			continue
		}
		if msg == "quit" {
			s.Offline(usr)
			return
		}
		s.DoMessage(usr, msg)
	}
}
