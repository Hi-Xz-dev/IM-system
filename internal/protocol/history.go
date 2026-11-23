package protocol

import "IM-system/internal/domain"

const HistoryResponseType = "history"

type HistoryResponse struct {
	Type     string           `json:"type"`
	RoomID   int64            `json:"room_id,omitempty"`
	PeerID   int64            `json:"peer_id,omitempty"`
	Messages []domain.Message `json:"messages"`
}
