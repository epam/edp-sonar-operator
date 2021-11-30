package sonar

import "github.com/pkg/errors"

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return string(e)
}

func IsErrNotFound(err error) bool {
	_, ok := errors.Cause(err).(ErrNotFound)
	return ok
}
