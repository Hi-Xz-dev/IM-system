package protocol

import (
	"encoding/json"
	"fmt"

	"IM-system/internal/domain"
)

func EncodeMessage(msg domain.Message) ([]byte, error) {
	//把 Go 对象转换成 JSON 格式的字节数据
	return json.Marshal(msg)

}

func DecodeMessage(data []byte) (domain.Message, error) {

	var msg domain.Message

	if err := json.Unmarshal(data, &msg); err != nil {
		return domain.Message{},
			fmt.Errorf("decode message: %w", err)
	}

	return msg, nil
}