package errors

type ResponseError struct {
	Error string `json:"message"`
}
