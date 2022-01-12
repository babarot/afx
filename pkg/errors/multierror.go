package errors

// Copied from https://github.com/hashicorp/go-multierror/

import (
	"fmt"
	"strings"
)

// Errors is an error type to track multiple errors. This is used to
// accumulate errors in cases and return them as a single "error".
type Errors []error

func (e *Errors) Error() string {
	format := func(text string) string {
		var s string
		lines := strings.Split(text, "\n")
		switch len(lines) {
		default:
			s += lines[0]
			for _, line := range lines[1:] {
				s += fmt.Sprintf("\n\t  %s", line)
			}
		case 0:
			s = (*e)[0].Error()
		}
		return s
	}

	if len(*e) == 1 {
		if (*e)[0] == nil {
			return ""
		}
		return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", format((*e)[0].Error()))
	}

	var points []string
	for _, err := range *e {
		if err == nil {
			continue
		}
		points = append(points, fmt.Sprintf("* %s", format(err.Error())))
	}

	return fmt.Sprintf(
		"%d errors occurred:\n\t%s\n\n",
		len(points), strings.Join(points, "\n\t"))
}

// Append is a helper function that will append more errors
// onto an Error in order to create a larger multi-error.
//
// If err is not a multierror.Error, then it will be turned into
// one. If any of the errs are multierr.Error, they will be flattened
// one level into err.
func (e *Errors) Append(errs ...error) {
	if e == nil {
		e = new(Errors)
	}
	for _, err := range errs {
		if err != nil {
			*e = append(*e, err)
		}
	}
}

// ErrorOrNil returns an error interface if this Error represents
// a list of errors, or returns nil if the list of errors is empty. This
// function is useful at the end of accumulation to make sure that the value
// returned represents the existence of errors.
func (e *Errors) ErrorOrNil() error {
	if e == nil {
		return nil
	}
	if len(*e) == 0 {
		return nil
	}

	return e
}
