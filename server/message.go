package server

import (
	"IM-system/internal/domain"
	"IM-system/internal/logger"
	"IM-system/internal/protocol"
	"IM-system/user"

	"fmt"
)

// 系统消息
func (s *Server) SendSystemMessage(u *user.User, content string) error {
	msg := domain.SystemMessage(content)
	data, err := protocol.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("encode system message: %w", err)
	}
	return u.SendMsg(string(data))
}

// 私聊消息
func (s *Server) SendPrivateMessage(sender *user.User, targets []*user.User, targetID int64, content string) error {
	msg := domain.NewMessage(domain.MessagePrivate, sender.ID, sender.Nickname, content)
	msg.To = targetID
	data, err := protocol.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("encode private message: %w", err)
	}
	for _, u := range targets {
		if err := u.SendMsg(string(data)); err != nil {
			logger.Log.Warn("send private message failed", "target_id", u.ID, "error", err)
			continue
		}
	}
	return nil
}

// 房间信息
func (s *Server) SendRoomMessage(sender *user.User, users []*user.User, roomID, content string) error {
	msg := domain.NewMessage(domain.MessageRoom, sender.ID, sender.Nickname, content)
	msg.RoomID = roomID
	data, err := protocol.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("encode room message: %w", err)
	}
	for _, u := range users {
		if err := u.SendMsg(string(data)); err != nil {
			logger.Log.Warn("send room message failed", "user_id", u.ID, "error", err)
		}
	}
	return nil
}

// BroadcastSystemMessage 全服广播（From=发送者信息）
func (s *Server) BroadcastSystemMessage(usr *user.User, content string) error {
	msg := domain.NewMessage(domain.MessageSystem, usr.ID, usr.Nickname, content)
	data, err := protocol.EncodeMessage(msg)
	if err != nil {
		return fmt.Errorf("encode broadcast message: %w", err)
	}
	s.Message <- string(data)
	return nil
}