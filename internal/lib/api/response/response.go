package response

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	InvalidRequest = "invalid request"

	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() *Response {
	return &Response{
		Status: StatusOK,
	}
}

func Error(msg string) *Response {
	return &Response{
		Status: StatusError,
		Error:  msg,
	}
}
