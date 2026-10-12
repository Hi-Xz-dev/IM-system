package server

import (
	"context"

	"IM-system/internal/logger"
)

func (s *Server) ListenPersistMessage() {
	for msg := range s.PersistQueue {
		if s.messageService == nil {
			continue
		}

		if err := s.messageService.Persist(
			context.Background(),
			msg,
		); err != nil {
			logger.Log.Error(
				"persist message failed",
				"message_type", msg.Type,
				"from_id", msg.From,
				"error", err,
			)
		}
	}
}
