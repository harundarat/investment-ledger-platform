package dto

type ErrorDetail struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ResponseMeta struct {
	IdempotencyReplayed bool `json:"idempotency_replayed"`
}

type Envelope struct {
	Message string         `json:"message,omitempty"`
	Data    any            `json:"data,omitempty"`
	Meta    *ResponseMeta  `json:"meta,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
}
