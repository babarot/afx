package diags

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type diagnostic struct {
	msg string
}

func (d diagnostic) Error() string {
	return d.msg
}

func New(msg string) error {
	return diagnostics{
		diagnostic{msg: msg},
	}
}

type diagnostics []error

type Error = diagnostics

var Verbose bool = false

func (ds diagnostics) Error() string {
	if len(ds) == 0 {
		return ""
	}

	err := &multierror.Error{
		ErrorFormat: MultiErrorFormat(),
	}

	if Verbose {
		for _, d := range ds {
			err = multierror.Append(err, d)
		}
		return err.Error()
	}

	err = multierror.Append(err, ds[len(ds)-1])
	return err.Error()
}

func (ds diagnostics) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			for _, err := range ds {
				pkgerr, ok := err.(interface{ Cause() error })
				if ok {
					fmt.Fprintf(s, "%+v\n\n", pkgerr)
				}
			}
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, ds.Error())
	case 'q':
		fmt.Fprintf(s, "%q", ds.Error())
	}
}

func (ds *diagnostics) Append(errs ...error) {
	if ds == nil {
		ds = new(diagnostics)
	}
	for _, err := range errs {
		if err != nil {
			*ds = append(*ds, err)
		}
	}
}

func (ds *diagnostics) ErrorOrNil() error {
	if ds == nil {
		return nil
	}
	if len(*ds) == 0 {
		return nil
	}
	return ds
}

func Wrap(err error, msg string) error {
	return errors.Wrap(err, msg)
}

func Wrapf(err error, format, msg string) error {
	// return errors.Wrapf(err, format, msg)
	switch e := err.(type) {
	case diagnostics:
		e.Append(errors.Wrapf(err, format, msg))
		return e
	default:
		return errors.Wrapf(err, format, msg)
	}
	// var ds diagnostics
	// ds.Append(err, errors.Wrapf(err, format, msg))
	// log.Printf("[DEBUG] ==================")
	// for _, d := range ds {
	// 	log.Printf("[DEBUG] %s", d)
	// }
	// return ds
}

// MultiErrorFormat provides a format for multierrors. This matches the default format, but if there
// is only one error we will not expand to multiple lines.
func MultiErrorFormat() multierror.ErrorFormatFunc {
	return func(es []error) string {
		format := func(text string) string {
			var s string
			lines := strings.Split(text, "\n")
			switch len(lines) {
			default:
				s += lines[0]
				for _, line := range lines[1:] {
					if line == "" {
						continue
					}
					s += fmt.Sprintf("\n\t  %s", line)
				}
			case 0:
				s = es[0].Error()
			}
			return s
		}

		if len(es) == 1 {
			if es[0] == nil {
				return ""
			}
			return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", format(es[0].Error()))
		}

		var points []string
		for _, err := range es {
			if err == nil {
				continue
			}
			points = append(points, fmt.Sprintf("* %s", format(err.Error())))
		}

		return fmt.Sprintf(
			"%d errors occurred:\n\t%s\n\n",
			len(points), strings.Join(points, "\n\t"))
	}
}
