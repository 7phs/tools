package common

import (
	"bytes"
	"github.com/pkg/errors"
)

type severalErrors struct {
	errs    []error
	message string
}

func (w *severalErrors) Error() string {
	buf := bytes.NewBufferString(w.message + ": ")

	for i, err := range w.errs {
		if i>0 {
			buf.WriteString("; ")
		}

		buf.WriteString(err.Error())
	}

	return buf.String()
}

func SeveralErrors(message string, listErrs ... error) error {
	errs := make([]error, 0, len(listErrs))

	for _, err := range listErrs {
		if err==nil {
			continue
		}

		switch e := err.(type) {
		case *severalErrors:
			errs = append(errs, e.errs...)
			errs = append(errs, errors.New(e.message))
		default:
			errs = append(errs, err)
		}
	}

	if len(errs)==0 {
		return nil
	}

	return &severalErrors{
		errs:    errs,
		message: message,
	}
}