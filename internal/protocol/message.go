package protocol

import (
	"encoding/json"

	"IM-system/internal/domain"
)


func EncodeMessage(msg domain.Message,) ([]byte,error){
	//把 Go 对象转换成 JSON 格式的字节数据
	return json.Marshal(msg)

}