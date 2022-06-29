package message

type ErrorMessage struct {
	ErrorCode int      `json:"errorCode"`
	Message   string   `json:"message"`
	Error     error    `json:"error"`
	Details   []string `json:"details,omitempty"`
}

type RequestData struct {
	Result bool `json:"result"`
}

func (r *RequestData) Validate() error {
	//TODO implement me
	//panic("implement me")
	return nil
}
