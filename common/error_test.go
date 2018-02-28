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

func TestSeveralErrors_SeveralErrors(t *testing.T) {
	var (
		err1 = errors.New("error 1")
		err2 = errors.New("error 2")
		err3 = SeveralErrors("error 3", err1, err2)
		err4 = errors.New("error 4")
		err5 = errors.New("error 5")
	)

	err := SeveralErrors("error 6", err3, err4, err5)

	expected := "error 6: error 1; error 2; error 3; error 4; error 5"

	if exist := err.Error(); strings.Compare(exist, expected) != 0 {
		t.Error("failed to create error with several errors. Got '", exist, "', but expected is '", expected, "'")
	}

}