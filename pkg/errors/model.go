package errors

type ResponseError struct {
	Message string `json:"error"`
	Details string `json:"details"`
}

func (r ResponseError) Error() string {
	return r.Message
}
