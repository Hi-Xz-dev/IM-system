package message

import (
	"context"
	"errors"
	"fmt"

	"IM-system/internal/domain"
	"IM-system/internal/model"
	"IM-system/internal/repository"
)

type MessageService struct {
	messageRepo *repository.MessageRepository
}

func NewMessageService(
	messageRepo *repository.MessageRepository,
) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
	}
}

// 把一条聊天消息保存到数据库
func (s *MessageService) Persist(
	ctx context.Context,
	msg domain.Message,
) error {
	dbMsg := &model.Message{
		Type:         string(msg.Type),
		FromID:       msg.From,
		FromNickname: msg.FromNickname,
		Content:      msg.Content,
		CreatedAt:    msg.Time,
		MessageID:    msg.ID,
	}

	switch msg.Type {

	case domain.MessagePrivate:
		toID := msg.To
		dbMsg.ToID = &toID

	case domain.MessageRoom:
		roomID := msg.RoomID
		roomName := msg.RoomName

		dbMsg.RoomID = &roomID
		dbMsg.RoomName = &roomName

	case domain.MessageSystem:
		// 系统消息如果只是普通系统通知，
		// to_id / room_id 都可以保持 nil

	case domain.MessageText:
		// 普通文本消息按后续业务定义处理
	}

	err := s.messageRepo.Create(ctx, dbMsg)
	if err == nil {
		return nil // 首次写入成功
	}

	if !errors.Is(err, repository.ErrDuplicateKey) {
		return fmt.Errorf("persist message: %w", err)
	}

	// 唯一键重复，查询已有消息
	existing, err := s.messageRepo.FindByMessageID(ctx, dbMsg.MessageID)
	if err != nil {
		return fmt.Errorf("find existing message: %w", err)
	}

	// 核对是不是同一条消息
	if existing.Type != dbMsg.Type ||
		existing.FromID != dbMsg.FromID ||
		existing.Content != dbMsg.Content ||
		!sameOptional(existing.ToID, dbMsg.ToID) ||
		!sameOptional(existing.RoomID, dbMsg.RoomID) {
		return fmt.Errorf("message ID conflict: %s", dbMsg.MessageID)
	}

	return nil // 同一条消息已经落库，重复消费也算成功
}

// 辅助函数 泛型函数
func sameOptional[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func toDomainMessages(
	dbMessages []model.Message,
) []domain.Message {

	messages := make(
		[]domain.Message,
		0,
		len(dbMessages),
	)

	for _, dbMsg := range dbMessages {
		msg := domain.Message{
			ID:           dbMsg.MessageID,
			Type:         domain.MessageType(dbMsg.Type),
			From:         dbMsg.FromID,
			FromNickname: dbMsg.FromNickname,
			Content:      dbMsg.Content,
			Time:         dbMsg.CreatedAt,
		}

		if dbMsg.ToID != nil {
			msg.To = *dbMsg.ToID
		}

		if dbMsg.RoomID != nil {
			msg.RoomID = *dbMsg.RoomID
		}

		if dbMsg.RoomName != nil {
			msg.RoomName = *dbMsg.RoomName
		}

		messages = append(messages, msg)
	}

	return messages
}

func (s *MessageService) GetRoomHistory(
	ctx context.Context,
	roomID int64,
	before *model.HistoryPosition,
) (HistoryPage, error) {
	dbMessages, err := s.messageRepo.FindByRoomID(ctx, roomID, historyPageSize+1, before)
	if err != nil {
		return HistoryPage{}, fmt.Errorf("get room history: %w", err)
	}
	return newHistoryPage(dbMessages), nil
}

func (s *MessageService) GetPrivateHistory(
	ctx context.Context,
	userID1, userID2 int64,
	before *model.HistoryPosition,
) (HistoryPage, error) {
	dbMessages, err := s.messageRepo.FindPrivateMessages(ctx, userID1, userID2, historyPageSize+1, before)
	if err != nil {
		return HistoryPage{}, fmt.Errorf("get private history: %w", err)
	}
	return newHistoryPage(dbMessages), nil
}
