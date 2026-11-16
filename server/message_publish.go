package server

import (
	"context"
	"fmt"
	"time"

	"IM-system/internal/domain"
)

func (s *Server) publishMessage(msg domain.Message) error {
	if s.isShuttingDown() {
		return fmt.Errorf("server is shutting down")
	}
	if s.messagePublisher == nil {
		return fmt.Errorf("message publisher is not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.messagePublisher.Publish(ctx, msg); err != nil {
		return fmt.Errorf("publish message: %w", err)
	}
	return nil
}
