package common

import (
	"bytes"
	"strings"
	"testing"
)

func TestMakeStream(t *testing.T) {
	expected := `line 1
line 2
line 3
`
	dst := bytes.NewBufferString("")

	stream := MakeStream(bytes.NewBufferString(expected))
	for str := range stream {
		dst.WriteString(str)
	}

	if exist := dst.String(); strings.Compare(exist, expected) != 0 {
		t.Error("failed to copy string using stream. Got '", exist, "', but expected is '", expected, "'")
	}

	if exist := ReadAll(MakeStream(bytes.NewBufferString(expected))); strings.Compare(exist, expected) != 0 {
		t.Error("failed to read all data from stream. Got '", exist, "', but expected is '", expected, "'")
	}
}

func TestMakeStream_Nil(t *testing.T) {
	if MakeStream(nil)!=nil {
		t.Error("failed to check nil reader")
	}
}
