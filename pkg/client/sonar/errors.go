package sonar

import (
	"errors"
	"fmt"
)

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return string(e)
}

func IsErrNotFound(err error) bool {
	errNotFound := ErrNotFound("")
	ok := errors.As(err, &errNotFound)
	return ok
}

type HTTPError struct {
	code    int
	message string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("status: %d, body: %s", e.code, e.message)
}

func IsHTTPErrorCode(err error, code int) bool {
	var httpError HTTPError
	ok := errors.As(err, &httpError)
	if !ok {
		return false
	}

	return httpError.code == code
}
