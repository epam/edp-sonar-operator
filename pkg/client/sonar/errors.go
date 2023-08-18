package sonar

import (
	"errors"
	"fmt"
	"net/http"
)

func IsErrNotFound(err error) bool {
	return IsHTTPErrorCode(err, http.StatusNotFound)
}

type HTTPError struct {
	code    int
	message string
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message string) HTTPError {
	return HTTPError{code: code, message: message}
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
