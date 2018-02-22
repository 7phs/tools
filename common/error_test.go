package common

import (
	"testing"
	"errors"
	"os"
	"strings"
)

func TestSeveralErrors(t *testing.T) {
	var (
		err1 = errors.New("error 1")
		err2 error
		err3 = errors.New("error 3")
		err4 error
		err5 = os.ErrExist
	)

	err := SeveralErrors("test errors", err1, err2, err3, err4, err5)

	expected := "test errors: error 1; error 3; file already exists"

	if exist := err.Error(); strings.Compare(exist, expected) != 0 {
		t.Error("failed to create error with several errors. Got '", exist, "', but expected is '", expected, "'")
	}
}

func TestSeveralErrors_Nil(t *testing.T) {
	var (
		err1 error
		err2 error
		err3 error
		err4 error
		err5 error
	)

	err := SeveralErrors("test errors", err1, err2, err3, err4, err5)

	if err!=nil {
		t.Error("failed to check nils errors. Got ", err)
	}
}