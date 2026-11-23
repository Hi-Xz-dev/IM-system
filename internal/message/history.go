package message

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"IM-system/internal/domain"
	"IM-system/internal/model"
)

const historyPageSize = 20

type HistoryPage struct {
	Messages   []domain.Message
	NextCursor string
}

func newHistoryPage(rows []model.Message) HistoryPage {
	var next string
	if len(rows) > historyPageSize {
		rows = rows[:historyPageSize]
		last := rows[len(rows)-1]
		value := last.CreatedAt.UTC().Format(time.RFC3339Nano) + "/" + strconv.FormatInt(last.ID, 10)
		next = base64.RawURLEncoding.EncodeToString([]byte(value))
	}
	return HistoryPage{Messages: toDomainMessages(rows), NextCursor: next}
}

var errInvalidHistoryCursor = errors.New("invalid history cursor")

func ParseHistoryCursor(cursor string) (*model.HistoryPosition, error) {
	//判断游标
	if cursor == "" {
		return nil, nil
	}
	if len(cursor) > 128 {
		return nil, errInvalidHistoryCursor
	}
	
	value, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, errInvalidHistoryCursor
	}
	// 检查解码错误
	timestamp, rowID, ok := strings.Cut(string(value), "/")
	if !ok {
		return nil, errInvalidHistoryCursor
	}
	createdAt, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil || createdAt.UTC().Year() < 1000 || createdAt.UTC().Year() > 9999 {
		return nil, errInvalidHistoryCursor
	}
	id, err := strconv.ParseInt(rowID, 10, 64)
	if err != nil || id <= 0 {
		return nil, errInvalidHistoryCursor
	}
	return &model.HistoryPosition{CreatedAt: createdAt.UTC(), ID: id}, nil
}
