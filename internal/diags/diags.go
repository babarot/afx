package diags

import (
	"fmt"
	"io"

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
	// return diagnostics{diagnostic{msg: msg}}
	return diagnostic{msg: msg}
}

type diagnostics []error

type Error = diagnostics

func (ds diagnostics) Error() string {
	var err error
	if len(ds) == 0 {
		return ""
	}
	for _, d := range ds {
		err = multierror.Append(err, d)
	}
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

	// var err *multierror.Error
	// for _, d := range *ds {
	// 	err = multierror.Append(err, d)
	// }
	// return err.ErrorOrNil()
}

func convert(err *multierror.Error) diagnostics {
	var ds diagnostics
	for _, e := range err.Errors {
		ds = append(ds, e)
	}
	return ds
}

func Wrap(err error, msg string) error {
	return errors.Wrap(err, msg)
}

func Wrapf(err error, format, msg string) error {
	return errors.Wrapf(err, format, msg)
}
