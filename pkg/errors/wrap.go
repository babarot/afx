package errors

import (
	"github.com/pkg/errors"
)

func New(messages ...string) error {
	var e Errors
	for _, message := range messages {
		e.Append(errors.New(message))
	}
	return e.ErrorOrNil()
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}
