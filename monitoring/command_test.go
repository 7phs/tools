package monitoring

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/7phs/tools/common"
)

func TestNewCoreCmd(t *testing.T) {
	cmd, err := NewCoreCmd(&testParameter{
		command: "echo",
		args:    []string{"hello", "world"},
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}
}

func TestCoreCmd_prepareStream(t *testing.T) {
	cmd := CoreCmd{}

	if stream, err := cmd.reader2Stream(nil, errors.New("test error")); stream != nil || err == nil {
		t.Error("failed to catch error")
	}

	if stream, err := cmd.reader2Stream(nil, nil); stream != nil || err == nil {
		t.Error("failed to catch nill reader")
	}

	if stream, err := cmd.reader2Stream(bytes.NewBufferString("hello"), nil); stream == nil || err != nil {
		t.Error("failed to make stream with ", err)
	}
}

func TestCoreCmd_Start(t *testing.T) {
	expected := "hello world\n"

	cmd, err := NewCoreCmd(&testParameter{
		command: "echo",
		args:    strings.Split(strings.TrimSpace(expected), " "),
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	var (
		prepare sync.WaitGroup
		wait    sync.WaitGroup
		exist   string
	)

	prepare.Add(1)
	wait.Add(1)
	go func() {
		prepare.Done()
		exist = common.ReadAll(cmd.StdOut())
		wait.Done()
	}()

	prepare.Wait()
	if err := cmd.Start(); err != nil {
		t.Error("failed to start command with", err)
		return
	}

	<-cmd.Wait()
	wait.Wait()

	if exist != expected {
		t.Error("failed to read command result. Got '", exist, "', but expected is '", expected, "'")
	}
}

func TestCoreCmd_StartErr(t *testing.T) {
	cmd, err := NewCoreCmd(&testParameter{
		command: "sleep",
		args:    []string{"2"},
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	if err := cmd.Start(); err != nil {
		t.Error("failed to start command with", err)
		return
	}

	if err := cmd.Start(); err == nil {
		t.Error("failed to catch error while start command")
	}

	cmd.Kill()
	<-cmd.Wait()
}

func TestCoreCmd_StdErr(t *testing.T) {
	unknownSh := fmt.Sprintf("unknown%d.sh", rand.Intn(9999))

	cmd, err := NewCoreCmd(&testParameter{
		command: "bash",
		args: []string{
			unknownSh,
			"123",
		},
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	var (
		prepare  sync.WaitGroup
		wait     sync.WaitGroup
		existOut string
		existErr string
	)

	prepare.Add(1)
	wait.Add(1)
	go func() {
		prepare.Done()
		existOut = common.ReadAll(cmd.StdOut())
		wait.Done()
	}()

	prepare.Add(1)
	wait.Add(1)
	go func() {
		prepare.Done()
		existErr = common.ReadAll(cmd.StdErr())
		wait.Done()
	}()

	prepare.Wait()
	if err := cmd.Start(); err != nil {
		t.Error("failed to start command with", err)
		return
	}

	<-cmd.Wait()
	wait.Wait()

	if existOut != "" {
		t.Error("unexpected out: '", existOut, "'")
	}

	if existErr == "" {
		t.Error("failed to catch stderr. Empty result")
	}

	if !strings.Contains(existErr, unknownSh) {
		t.Error("unexpected stderr: '", existErr, "'; without '", unknownSh, "'")
	}
}

func TestCoreCmd_Kill(t *testing.T) {
	timeout := 5

	cmd, err := NewCoreCmd(&testParameter{
		command: "sleep",
		args:    []string{strconv.Itoa(timeout)},
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	if err := cmd.Start(); err != nil {
		t.Error("failed to start command with", err)
		return
	}

	var mErr error
	cmd.Monitoring(func(monitoringErr error) {
		mErr = monitoringErr
	})

	var (
		wait     sync.WaitGroup
		duration = 0 * time.Second
	)

	wait.Add(1)
	go func() {
		start := time.Now()
		<-cmd.Wait()
		duration = time.Now().Sub(start)

		wait.Done()
	}()

	if err := cmd.Kill(); err != nil {
		t.Error("failed to kill command with", err)
	}

	wait.Wait()

	if duration == 0 || duration >= time.Duration(timeout)*time.Second {
		t.Error("failed to kill the command in time. Got duration", duration, ", but it should be by zero")
	}
}

func TestCoreCmd_KillErr(t *testing.T) {
	//========== 1
	cmd, err := NewCoreCmd(&testParameter{
		command: "sleep",
		args:    []string{"5"},
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	if err := cmd.Kill(); err == nil {
		t.Error("failed to cacth error")
	}

	//========== 2
	cmd, err = NewCoreCmd(&testParameter{
		command: "sleep",
		args:    []string{"1"},
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	if err := cmd.Start(); err != nil {
		t.Error("failed to start command with", err)
		return
	}

	if err := cmd.Kill(); err != nil {
		t.Error("failed to kill the command with", err)
	}

	time.Sleep(200 * time.Millisecond)

	if err := cmd.Kill(); err == nil {
		t.Error("failed to kill the command second time without an error")
	}

}

func TestCoreCmd_Monitoring(t *testing.T) {
	expected := "hello world"

	cmd, err := NewCoreCmd(&testParameter{
		command: "echo",
		args:    strings.Split(expected, " "),
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	if err := cmd.Start(); err != nil {
		t.Error("failed to start command with", err)
		return
	}

	var wait sync.WaitGroup

	wait.Add(1)
	go func() {
		cmd.Monitoring(func(err error) {
			wait.Done()
		})
	}()

	wait.Wait()

	if exist := strings.TrimSpace(common.ReadAll(cmd.StdOut())); strings.Compare(expected, exist) != 0 {
		t.Error("unexpected output of the command. Got '", exist, "', but expected is '", expected, "'")
	}
}

func TestCoreCmd_MonitoringErr(t *testing.T) {
	expected := "hello world"

	cmd, err := NewCoreCmd(&testParameter{
		command: "echo",
		args:    strings.Split(expected, " "),
	})
	if cmd == nil || err != nil {
		t.Error("failed to create command with", err)
		return
	}

	err = nil
	cmd.Monitoring(func(monitoringErr error) {
		err = monitoringErr
	})

	if err == nil {
		t.Error("failed to check error while start monitoring with", err)
		return
	}
}
