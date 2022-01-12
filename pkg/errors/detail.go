package errors

import (
	"fmt"
)

// Detail is
type Detail struct {
	Head    string
	Summary string
	Details []string
}

func (e Detail) Error() string {
	var msg string

	if e.Head == "" && e.Summary == "" {
		panic("cannot use detailed error type")
	}

	if e.Head == "" {
		msg = e.Summary
	} else {
		msg = fmt.Sprintf("%s: %s", e.Head, e.Summary)
	}

	if len(e.Details) > 0 {
		msg += "\n\n"
		for _, line := range e.Details {
			msg += fmt.Sprintf("\t%s\n", line)
		}
	}

	return msg
}
