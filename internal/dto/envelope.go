package dto

type ErrorResponse struct {
	Message string `json:"message"`
}

type Envelope struct {
	Message string         `json:"message,omitempty"`
	Data    any            `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
}
