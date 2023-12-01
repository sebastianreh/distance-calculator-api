package exceptions

type NotFoundException interface {
	Error() string
	IsNotFoundError() bool
}

type notFoundException struct {
	ErrMessage string
}

func (exception *notFoundException) Error() string {
	return exception.ErrMessage
}

func (exception *notFoundException) IsNotFoundError() bool {
	return true
}

func NewNotFoundException(message string) NotFoundException {
	return &notFoundException{ErrMessage: message}
}
