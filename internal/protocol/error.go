package protocol

import "IM-system/internal/domain"

type ErrorResponse struct {
	Type    string             `json:"type"`
	Command domain.CommandType `json:"command"`
	Code    string             `json:"code"`
	Message string             `json:"message"`
}
