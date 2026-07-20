package handler

import "github.com/harundarat/investment-ledger-platform/internal/dto"

func success(data any, message string) *dto.Envelope {
	return &dto.Envelope{Data: data, Message: message}
}

func fail(code, message string, details ...dto.ErrorDetail) *dto.Envelope {
	return &dto.Envelope{Error: &dto.ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}}
}
