package errors

// https://github.com/upspin/upspin/tree/master/errors
// https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
// https://blog.golang.org/errors-are-values

import (
	"bytes"
	"fmt"
	"log"
	"runtime"

	"github.com/hashicorp/hcl/v2"
)

var (
	_ error = (*Error)(nil)
)

// Error is the type that implements the error interface.
// It contains a number of fields, each of different type.
// An Error value may leave some values unset.
type Error struct {
	Err         error
	Files       map[string]*hcl.File
	Diagnostics hcl.Diagnostics
}

func (e *Error) Error() string {
	switch {
	case e.Diagnostics.HasErrors():
		if len(e.Files) == 0 {
			return e.Diagnostics.Error()
		}
		var b bytes.Buffer
		wr := hcl.NewDiagnosticTextWriter(
			&b,      // writer to send messages to
			e.Files, // the parser's file cache, for source snippets
			100,     // wrapping width
			true,    // generate colored/highlighted output
		)
		wr.WriteDiagnostics(e.Diagnostics)
		return b.String()
	case e.Err != nil:
		return e.Err.Error()
	}
	return ""
}

func flatten(diags hcl.Diagnostics) hcl.Diagnostics {
	m := make(map[string]bool)
	var d hcl.Diagnostics
	for _, diag := range diags {
		err := diags.Error()
		if !m[err] {
			m[err] = true
			d = append(d, diag)
		}
	}
	return d
}

// New is
func New(args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			// Someone might accidentally call us with a user or path name
			// that is not of the right type. Take care of that and log it.
			// if strings.Contains(arg, "@") {
			// 	_, file, line, _ := runtime.Caller(1)
			// 	log.Printf("errors.E: unqualified type for %q from %s:%d", arg, file, line)
			// 	if strings.Contains(arg, "/") {
			// 		if e.Path == "" { // Don't overwrite a valid path.
			// 			e.Path = upspin.PathName(arg)
			// 		}
			// 	} else {
			// 		if e.User == "" { // Don't overwrite a valid user.
			// 			e.User = upspin.UserName(arg)
			// 		}
			// 	}
			// 	continue
			// }
			e.Err = &errorString{arg}
		case *Error:
			// Make a copy
			copy := *arg
			e.Err = &copy
		case hcl.Diagnostics:
			e.Diagnostics = flatten(arg)
		case map[string]*hcl.File:
			e.Files = arg
		case error:
			e.Err = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			log.Printf("errors.E: bad call from %s:%d: %v", file, line, args)
			return Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}

	// prev, ok := e.Err.(*Error)
	// if !ok {
	// 	return e
	// }
	//
	// // The previous error was also one of ours. Suppress duplications
	// // so the message won't contain the same kind, file name or user name
	// // twice.
	// if prev.Path == e.Path {
	// 	prev.Path = ""
	// }
	// if prev.User == e.User {
	// 	prev.User = ""
	// }
	// if prev.Kind == e.Kind {
	// 	prev.Kind = Other
	// }
	// // If this error has Kind unset or Other, pull up the inner one.
	// if e.Kind == Other {
	// 	e.Kind = prev.Kind
	// 	prev.Kind = Other
	// }

	if len(e.Diagnostics) == 0 && e.Err == nil {
		return nil
	}

	return e
}

// Recreate the errors.New functionality of the standard Go errors package
// so we can create simple text errors when needed.

// New returns an error that formats as the given text. It is intended to
// be used as the error-typed argument to the E function.
// func New(text string) error {
// 	return &errorString{text}
// }

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Errorf is equivalent to fmt.Errorf, but allows clients to import only this
// package for all error handling.
func Errorf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}
