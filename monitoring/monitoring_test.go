package monitoring

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
	"sync"
)

func TestMonitoringCommand_String(t *testing.T) {
	expected := "cmdStart"
	if exist := cmdStart.String(); strings.Compare(exist, expected) != 0 {
		t.Error("failed to convert to string cmdStart. Got", exist)
	}

	unknown := 1024
	expected = "monitoringCommand(" + strconv.Itoa(unknown) + ")"
	if exist := monitoringCommand(unknown).String(); strings.Compare(exist, expected) != 0 {
		t.Error("failed to convert to string unknown number ", unknown, ". Got", exist, ", but expected is", expected)
	}
}

func TestMonitoringStage_String(t *testing.T) {
	expected := "execFinish"
	if exist := execFinish.String(); strings.Compare(exist, expected) != 0 {
		t.Error("failed to convert to string execShutdown. Got", exist)
	}

	unknown := 1024
	expected = "monitoringStage(" + strconv.Itoa(unknown) + ")"
	if exist := monitoringStage(unknown).String(); strings.Compare(exist, expected) != 0 {
		t.Error("failed to convert to string unknown number ", unknown, ". Got", exist, ", but expected is", expected)
	}
}

func TestNewMonitoring(t *testing.T) {
	monitoring := NewMonitoring(&testParameter{
		command: "echo",
		args:    []string{"hello", "world"},
	})
	if monitoring == nil {
		t.Error("failed to create command monitoring with")
		return
	}
}

func TestMonitoringWorkflow(t *testing.T) {
	monitoring := NewMonitoring(&testParameter{
		command:       "echo",
		args:          []string{"hello", "world"},
		parallelCount: 15,
	})

	if monitoring == nil {
		t.Error("failed to create command monitoring with")
		return
	}

	monitoring.Start(context.Background())

	fmt.Println("monitoring.Start:finish")

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	monitoring.Stop(context.Background())

	monitoring.Wait()

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to stop command monitoring with", err)
	}
}

func TestMonitoringWorkflow_Stress(t *testing.T) {
	{
		monitoring := NewMonitoring(&testParameter{
			command:       "echo",
			args:          []string{"hello", "world"},
			parallelCount: 15,
		})

		if monitoring == nil {
			t.Error("failed to create command monitoring with")
			return
		}

		go func() {
			monitoring.Start(context.Background())
		}()
		go func() {
			monitoring.Stop(context.Background())
		}()
		monitoring.Wait()
	}

	{
		monitoring := NewMonitoring(&testParameter{
			command: "echo",
			args:    []string{"hello", "world"},
		})

		if monitoring == nil {
			t.Error("failed to create command monitoring with")
			return
		}

		go func() {
			monitoring.Stop(context.Background())
		}()
		monitoring.Wait()
	}
}

func TestMonitoring_Waiting(t *testing.T) {
	expectedFrom := 0.35
	expectedTill := 1 * time.Second

	monitoring := NewMonitoring(&testParameter{
		command:     "sleep",
		args:        []string{strconv.FormatFloat(expectedFrom, 'f', 2, 64)},
		runningMode: RunOnce,
	})

	ctx, _ := context.WithTimeout(context.Background(), expectedTill)
	monitoring.Start(ctx)
	fmt.Println("start")
	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	start := time.Now()
	monitoring.Wait()
	exist := time.Since(start)

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to stop command monitoring with", err)
	}

	if from := time.Duration(expectedFrom * 1000); exist < from || exist > expectedTill {
		t.Error("failed to done command. Executed ", exist, ", but should be from ", from, " till ", expectedTill)
	}
}

func TestMonitoring_Timeout(t *testing.T) {
	expectedFrom := 5
	expectedTill := 100 * time.Millisecond

	monitoring := NewMonitoring(&testParameter{
		command:     "sleep",
		args:        []string{strconv.Itoa(expectedFrom)},
		runningMode: RunOnce,
	})

	ctx, _ := context.WithTimeout(context.Background(), expectedTill)
	start := time.Now()
	monitoring.Start(ctx)
	exist := time.Since(start)

	if err := monitoring.HasError(); err == nil {
		t.Error("failed to check timeout for start command monitoring")
	}

	monitoring.Wait()

	if exist < expectedTill || exist > expectedTill*2 {
		t.Error("failed to done command. Executed ", exist, ", but should be from ", expectedTill, " till ", expectedTill*2)
	}
}

func TestMonitoring_SilenceExit(t *testing.T) {
	monitoring := NewMonitoring(&testParameter{
		command:     "echo",
		args:        []string{"hello", "world", ">", "nul"},
		runningMode: RunOnce,
	})

	monitoring.Start(context.Background())

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	monitoring.Wait()

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to run command monitoring with", err)
	}
}

func TestMonitoring_Error(t *testing.T) {
	unknownSh := fmt.Sprintf("unknown%d.sh", rand.Intn(9999))

	monitoring := NewMonitoring(&testParameter{
		command:     "bash",
		args:        []string{unknownSh, "123"},
		runningMode: RunOnce,
	})

	monitoring.Start(context.Background())

	if err := monitoring.HasError(); err == nil {
		t.Error("failed to catch error command")
	}

	monitoring.Stop(context.Background())

	monitoring.Wait()

	if err := monitoring.HasError(); err == nil {
		t.Error("failed to catch error command after stop")
	}
}

func TestMonitoring_Kill(t *testing.T) {
	monitoring := NewMonitoring(&testParameter{
		command:       "echo",
		args:          []string{"hello world"},
		runningMode:   RepeatInfinity,
		parallelCount: 15,
	})

	if monitoring == nil {
		t.Error("failed to create command monitoring with")
		return
	}

	monitoring.Start(context.Background())

	fmt.Println("monitoring.Start:finish")

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	time.Sleep(200 * time.Millisecond)

	monitoring.Kill(context.Background())

	fmt.Println("monitoring.Kill:finish")

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to kill command monitoring with", err)
	}

	monitoring.Start(context.Background())

	fmt.Println("monitoring.Start:2:finish")

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	time.Sleep(200 * time.Millisecond)

	monitoring.Kill(context.Background())

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to kill command monitoring with", err)
	}

	monitoring.Start(context.Background())

	fmt.Println("monitoring.Start:2:finish")

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	monitoring.Stop(context.Background())

	monitoring.Wait()

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to stop command monitoring with", err)
	}
}

func TestMonitoring_StartCancellation(t *testing.T) {
	monitoring := NewMonitoring(&testParameter{
		command:       "sleep",
		args:          []string{"5"},
		parallelCount: 15,
	})

	if monitoring == nil {
		t.Error("failed to create command monitoring with")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	var (
		waitToStart sync.WaitGroup
		waitToFinish sync.WaitGroup
	)
	waitToStart.Add(1)
	waitToFinish.Add(1)
	go func() {
		waitToStart.Done()

		monitoring.Start(ctx)

		waitToFinish.Done()
	}()
	waitToStart.Wait()

	cancel()

	waitToFinish.Wait()

	fmt.Println("monitoring.Start:finish")

	if err := monitoring.HasError(); err == nil {
		t.Error("failed to catch cancelation in start command monitoring with", err)
	}

	monitoring.Stop(context.Background())

	monitoring.Wait()

	if err := monitoring.HasError(); err == nil {
		t.Error("failed to catch error stop command monitoring with", err)
	}
}


func TestMonitoring_KillCancellation(t *testing.T) {
	monitoring := NewMonitoring(&testParameter{
		command:       "echo",
		args:          []string{"hello world"},
		parallelCount: 15,
	})

	if monitoring == nil {
		t.Error("failed to create command monitoring with")
		return
	}

	monitoring.Start(context.Background())

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var (
		waitToStart sync.WaitGroup
		waitToFinish sync.WaitGroup
	)
	waitToStart.Add(1)
	waitToFinish.Add(1)
	go func() {
		waitToStart.Done()

		monitoring.Kill(ctx)

		waitToFinish.Done()
	}()
	waitToStart.Wait()

	cancel()

	waitToFinish.Wait()

	if err := monitoring.HasError(); err == nil {
		t.Error("failed to catch cancelation in kill command monitoring with", err)
	}

	monitoring.Stop(context.Background())

	monitoring.Wait()
}


func TestMonitoring_StopCancellation(t *testing.T) {
	monitoring := NewMonitoring(&testParameter{
		command:       "echo",
		args:          []string{"hello world"},
		parallelCount: 15,
	})

	if monitoring == nil {
		t.Error("failed to create command monitoring with")
		return
	}

	monitoring.Start(context.Background())

	if err := monitoring.HasError(); err != nil {
		t.Error("failed to start command monitoring with", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var (
		waitToStart sync.WaitGroup
		waitToFinish sync.WaitGroup
	)
	waitToStart.Add(1)
	waitToFinish.Add(1)
	go func() {
		waitToStart.Done()

		monitoring.Stop(ctx)

		waitToFinish.Done()
	}()
	waitToStart.Wait()

	cancel()

	waitToFinish.Wait()

	monitoring.Wait()

	if err := monitoring.HasError(); err == nil {
		t.Error("failed to catch cancelation in stop command monitoring with", err)
	}
}

//func TestMonitoring_DoubleStart(t *testing.T) {
//
//}
