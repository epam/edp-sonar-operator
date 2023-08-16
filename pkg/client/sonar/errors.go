package sonar

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

type NotFoundError string

func (e NotFoundError) Error() string {
	return string(e)
}

func IsErrNotFound(err error) bool {
	errNotFound := NotFoundError("")
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
