package monitoring

import (
	"testing"
	"context"
	"strings"
)

func TestCommandCtx_String(t *testing.T) {
	cmd, _ := CommandCtx(context.Background(), cmdStart)

	expected := "context.Background.CommandCtx(cmdStart)"
	exist := cmd.(*commandCtx).String()
	if strings.Compare(exist, expected) != 0 {
		t.Error("failed to stringify command context. Got '", exist, "', but expecte is '", expected, "'")
	}
}

func TestCommandCtx_Value(t *testing.T) {
	parentName := "prop1"
	parentExpected := "value1"
	parentCtx := context.WithValue(context.Background(), parentName, parentExpected)

	parentName2 := "prop2"
	parentExpected2 := "value2"
	parentCtx2 := context.WithValue(parentCtx, parentName2, parentExpected2)

	expected := cmdStart
	cmd, _ := CommandCtx(parentCtx2, expected)

	if exist := cmd.Value("command").(monitoringCommand); exist != expected {
		t.Error("failed to get 'command' value. Got", exist, ", but expected", expected)
	}

	if exist := cmd.Value(parentName).(string); exist != parentExpected {
		t.Error("failed to get '", parentName, "' value. Got", exist, ", but expected", parentExpected)
	}

	if exist := cmd.Value(parentName2).(string); exist != parentExpected2 {
		t.Error("failed to get '", parentName2, "' value. Got", exist, ", but expected", parentExpected2)
	}
}
