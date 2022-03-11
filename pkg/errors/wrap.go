package errors

import (
	"github.com/pkg/errors"
)

func New(message string) error {
	var e Errors
	e.Append(errors.New(message))
	return e.ErrorOrNil()
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}
