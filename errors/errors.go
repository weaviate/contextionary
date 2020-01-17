package errors

import "fmt"

// InvalidUserInput indicates a client-side error
type InvalidUserInput struct {
	msg string
}

func (e InvalidUserInput) Error() string {
	return e.msg
}

// NewInvalidUserInput with Errorf signature
func NewInvalidUserInputf(format string, args ...interface{}) InvalidUserInput {
	return InvalidUserInput{msg: fmt.Sprintf(format, args...)}
}

// Internal indicates something went wrong during processing
type Internal struct {
	msg string
}

func (e Internal) Error() string {
	return e.msg
}

// NewInternal with Errorf signature
func NewInternalf(format string, args ...interface{}) Internal {
	return Internal{msg: fmt.Sprintf(format, args...)}
}

// NotFound indicates the desired resource doesn't exist
type NotFound struct {
	msg string
}

func (e NotFound) Error() string {
	return e.msg
}

// NewNotFound with Errorf signature
func NewNotFoundf(format string, args ...interface{}) NotFound {
	return NotFound{msg: fmt.Sprintf(format, args...)}
}
