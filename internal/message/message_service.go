package message

import (
	"context"
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

	if err := s.messageRepo.Create(ctx, dbMsg); err != nil {
		return fmt.Errorf(
			"persist message: %w",
			err,
		)
	}

	return nil
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
	limit int,
) ([]domain.Message, error) {

	dbMessages, err := s.messageRepo.FindByRoomID(
		ctx,
		roomID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get room history: %w",
			err,
		)
	}

	return toDomainMessages(dbMessages), nil
}

func (s *MessageService) GetPrivateHistory(
	ctx context.Context,
	userID1 int64,
	userID2 int64,
	limit int,
) ([]domain.Message, error) {

	dbMessages, err := s.messageRepo.FindPrivateMessages(
		ctx,
		userID1,
		userID2,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get private history: %w",
			err,
		)
	}

	return toDomainMessages(dbMessages), nil
}
