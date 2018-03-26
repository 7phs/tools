//go:generate stringer -type=monitoringCommand
//go:generate stringer -type=monitoringStage

package monitoring

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/7phs/tools/common"
	"github.com/pkg/errors"
)

type monitoringCommand int

type monitoringStage int

func (o monitoringStage) Int32() int32 {
	return int32(o)
}

func (o monitoringCommand) Int32() int32 {
	return int32(o)
}

const (
	execStopped monitoringStage = iota
	execStarting
	execMonitoring
	execFinish

	cmdStart monitoringCommand = iota
	cmdKill
	cmdStop
)

const (
	RepeatInfinity int32 = 0 - iota
	RunOnce
)

type MonitoringParameter interface {
	Parameter

	RunningMode() int32
	ParallelCount() int32
	StdErrIsOk() bool
	CheckStartLine() func(string) bool
}

type Monitoring struct {
	MonitoringParameter

	cmd []*commandState
	err error

	stage int32

	commandFlow chan context.Context
	commandWait sync.WaitGroup

	monitoring      chan interface{}
	monitoringState bool
	monitoringWait  sync.WaitGroup
}

func NewMonitoring(parameter MonitoringParameter) *Monitoring {
	return (&Monitoring{
		MonitoringParameter: parameter,

		commandFlow: make(chan context.Context),
	}).init()
}

func (o *Monitoring) init() *Monitoring {
	atomic.StoreInt32(&o.stage, execStopped.Int32())

	o.commandWait.Add(1)
	go o.commandFlowExecution()

	return o
}

func (o *Monitoring) Start(ctx context.Context) {
	if wait := o.runCommand(ctx, cmdStart); wait != nil {
		<-wait.Done()
	}
}

func (o *Monitoring) Kill(ctx context.Context) {
	if wait := o.runCommand(ctx, cmdKill); wait != nil {
		<-wait.Done()
	}
}

func (o *Monitoring) Stop(ctx context.Context) {
	if wait := o.runCommand(ctx, cmdStop); wait != nil {
		<-wait.Done()
	}
}

func (o *Monitoring) runCommand(ctx context.Context, command monitoringCommand) context.Context {
	ctx, _ = CommandCtx(ctx, command)
	o.commandFlow <- ctx

	return ctx
}

func (o *Monitoring) Wait() {
	o.commandWait.Wait()
}

func (o *Monitoring) catchError(err error) error {
	if err == nil {
		return err
	}

	o.err = err

	return err
}

func (o *Monitoring) HasError() error {
	return o.err
}

func (o *Monitoring) commandFlowExecution() {
	var (
		completely = false
	)

	for command := range o.commandFlow {
		if completely {
			if ctx, ok := command.(*commandCtx); ok {
				ctx.done(errors.New("monitoring is already stopped"))
			}

			continue
		}

		finish, _ := o.commandExecution(command)
		if finish && !completely {
			o.commandWait.Done()
			completely = true
		}
	}
}

func (o *Monitoring) commandExecution(ctx context.Context) (finish bool, err error) {
	var (
		command = ctx.Value("command").(monitoringCommand)
		doneErr error
	)
	ctx, done := CommandCtx(ctx, command)
	defer func() {
		done(doneErr)
	}()

	switch command {
	case cmdStart:
		finish, err = o.commandStart(ctx)

	case cmdKill:
		finish, err = o.commandKill(ctx)

	case cmdStop:
		finish, err = o.commandStop(ctx)

	default:
		err = errors.New(fmt.Sprint("unknown command:", command))
	}

	o.catchError(err)

	doneErr = err

	return
}

func (o *Monitoring) commandStart(ctx context.Context) (finish bool, err error) {
	if !atomic.CompareAndSwapInt32(&o.stage, execStopped.Int32(), execStarting.Int32()) {
		return false, errors.New("failed to start execute command start - already started")
	}

	if o.cmd == nil || len(o.cmd) == 0 {
		parallelCount := o.ParallelCount()
		if parallelCount <= 0 {
			parallelCount = 1
		}

		o.cmd = make([]*commandState, parallelCount)
	}

	for i := range o.cmd {
		o.cmd[i], err = CommandState(o)
		if err != nil {
			return true, errors.Wrapf(err, "failed to create command for monitoring")
		}
	}

	wait := make(chan error)
	for _, cmd := range o.cmd {
		go cmd.Run(ctx, wait)
	}

	for i := 0; i < len(o.cmd); i++ {
		if cmdErr := <-wait; cmdErr != nil {
			err = common.SeveralErrors("failed to start command:", err, cmdErr)
		}
	}

	// shutdown on failure
	if err == nil {
		err = o.startMonitoring()
	} else {
		o.killAll()
	}

	return err != nil, err
}

func (o *Monitoring) killAll() {
	var wait sync.WaitGroup

	if o.monitoringState {
		close(o.monitoring)
		o.monitoringState = false
	}

	for _, cmd := range o.cmd {
		wait.Add(1)
		go func(cmd *commandState) {
			if cmd != nil && !cmd.IsExited() {
				cmd.Kill()
			}

			wait.Done()
		}(cmd)
	}

	wait.Wait()
}

func (o *Monitoring) startMonitoring() error {
	if !atomic.CompareAndSwapInt32(&o.stage, execStarting.Int32(), execMonitoring.Int32()) {
		return errors.New("failed to start monitoring - process wasn't started yet")
	}

	o.monitoring = make(chan interface{})
	o.monitoringState = true

	for _, cmd := range o.cmd {
		go o.monitoringProcess(cmd)
	}

	return nil
}

func (o *Monitoring) commandKill(ctx context.Context) (finish bool, err error) {
	if !atomic.CompareAndSwapInt32(&o.stage, execMonitoring.Int32(), execStopped.Int32()) {
		return false, errors.New("failed to execute command kill - already stopped")
	}

	wait := make(chan interface{})

	go func() {
		o.killAll()

		close(wait)
	}()

	select {
	case <-wait:

	case <-ctx.Done():
		err = errors.New("user cancel")
	}

	return true, err
}

func (o *Monitoring) commandStop(ctx context.Context) (finish bool, err error) {
	atomic.StoreInt32(&o.stage, execFinish.Int32())

	wait := make(chan interface{})

	go func() {
		o.killAll()

		close(wait)
	}()

	select {
	case <-wait:

	case <-ctx.Done():
		err = errors.New("user cancel")
	}

	return true, err
}

func (o *Monitoring) monitoringProcess(cmd *commandState) {
	repeat := o.RunningMode()
	wait := make(chan error)

	for {
		select {
		case <-cmd.Wait():

		case <-o.monitoring:
			return
		}

		switch {
		case repeat == RepeatInfinity, cmd.RunCount() <= repeat:
			cmd.Init(o)
			cmd.Run(context.Background(), wait)
			if o.catchError(<-wait) != nil {
				o.Stop(context.Background())
				return
			}

		case repeat == RunOnce, cmd.RunCount() > repeat:
			o.Stop(context.Background())
			return
		}
	}
}
