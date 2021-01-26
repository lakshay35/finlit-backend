package errors

// Error object
type Error struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (error Error) Error() string {
	return error.Message
}
