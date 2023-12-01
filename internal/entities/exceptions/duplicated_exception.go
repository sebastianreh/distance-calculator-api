package exceptions

type DuplicatedException interface {
	Error() string
	IsDuplicatedError() bool
}

type duplicatedException struct {
	ErrMessage string
}

type Causes struct {
	Code string `json:"code"`
}

func (exception *duplicatedException) Error() string {
	return exception.ErrMessage
}

func (exception *duplicatedException) IsDuplicatedError() bool {
	return true
}

func NewDuplicatedException(message string) DuplicatedException {
	return &duplicatedException{ErrMessage: message}
}
