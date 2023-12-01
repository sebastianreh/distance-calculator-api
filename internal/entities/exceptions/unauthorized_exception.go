package exceptions

type UnauthorizedException interface {
	Error() string
	IsUnauthorizedExceptionError() bool
}

type unauthorizedException struct {
	ErrMessage string
}

func (exception *unauthorizedException) Error() string {
	return exception.ErrMessage
}

func (exception *unauthorizedException) IsUnauthorizedExceptionError() bool {
	return true
}

func NewUnauthorizedException(message string) UnauthorizedException {
	return &unauthorizedException{ErrMessage: message}
}
