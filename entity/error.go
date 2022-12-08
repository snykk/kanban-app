package entity

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(msg string) ErrorResponse {
	return ErrorResponse{
		Error: msg,
	}
}
