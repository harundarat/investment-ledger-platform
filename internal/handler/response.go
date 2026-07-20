package handler

import "github.com/harundarat/investment-ledger-platform/internal/dto"

func success(data any, message string) *dto.Envelope {
	return &dto.Envelope{Data: data, Message: message}
}

func successWithMeta(data any, message string, meta *dto.ResponseMeta) *dto.Envelope {
	return &dto.Envelope{Data: data, Message: message, Meta: meta}
}

func fail(code, message string, details ...dto.ErrorDetail) *dto.Envelope {
	return &dto.Envelope{Error: &dto.ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}}
}
