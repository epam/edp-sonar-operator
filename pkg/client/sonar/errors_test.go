package sonar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsErrNotFoundFalse(t *testing.T) {
	assert.False(t, IsErrNotFound(nil))
}

func TestIsHTTPErrorCodeFalse(t *testing.T) {
	assert.False(t, IsHTTPErrorCode(nil, http.StatusOK))
}

func TestIsHTTPErrorCode(t *testing.T) {
	httpError := HTTPError{
		code:    http.StatusOK,
		message: "",
	}
	assert.True(t, IsHTTPErrorCode(httpError, http.StatusOK))
}
