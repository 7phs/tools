package monitoring

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/7phs/tools/common"
	"github.com/pkg/errors"
)

type commandState struct {
	CoreCmd

	stdErrIsOk     bool
	checkStartLine func(string) bool
	runCount       int32
}

func CommandState(parameter MonitoringParameter) (*commandState, error) {
	cmdState := &commandState{}
	err := cmdState.Init(parameter)

	return cmdState, err
}

func (o *commandState) Init(parameter MonitoringParameter) error {
	_, err := o.init(parameter).prepare()

	o.checkStartLine = parameter.CheckStartLine()
	o.stdErrIsOk = parameter.StdErrIsOk()

	return err
}

func (o *commandState) RunCount() int32 {
	return atomic.LoadInt32(&o.runCount)
}

func (o *commandState) Run(ctx context.Context, wait chan<- error) {
	firstLine, err := o.startCommand(ctx)

	if err == nil && !o.checkLine(firstLine) {
		err = o.waitStartLine(ctx)
	}

	wait <- err
}

func (o *commandState) checkLine(line string) bool {
	if o.checkStartLine == nil {
		return true
	}

	return o.checkStartLine(line)
}

func (o *commandState) startCommand(ctx context.Context) (line string, err error) {
	var (
		ok bool
	)

	err = o.Start()
	if err != nil {
		err = errors.Wrapf(err, "failed to start command for monitoring")
	} else {
		atomic.AddInt32(&o.runCount, 1)

		select {
		case <-common.DeadLineTimer(ctx.Deadline()).C:
			err = errors.New("timeout error")

			// check for std out and unexpected close stdOut and check stdErr
		case line, ok = <-o.StdOut():
			if !ok {
				if errMsg := common.ReadAll(o.StdErr()); len(errMsg) > 0 {
					err = errors.New(errMsg)
				}
			}

		case line, ok = <-o.StdErr():
			if !ok {
				if errMsg := common.ReadAll(o.StdErr()); len(errMsg) > 0 {
					err = errors.New(line + errMsg)
				}
			} else if !o.stdErrIsOk {
				err = errors.New(line)
			}

		case <-ctx.Done():
			err = errors.New("user cancel")
		}
	}

	return
}

func (o *commandState) waitStartLine(ctx context.Context) (err error) {
	var (
		line          string
		openedChannel = true
		foundLine     bool
	)

	for err == nil && openedChannel && !foundLine {
		select {
		case <-common.DeadLineTimer(ctx.Deadline()).C:
			err = errors.New("timeout error")

			// check for std out and unexpected close stdOut and check stdErr
		case line, openedChannel = <-o.StdOut():
			if !openedChannel {
				if errMsg := common.ReadAll(o.StdErr()); len(errMsg) > 0 {
					err = errors.New(errMsg)
				}
			} else {
				foundLine = o.checkLine(line)
			}

		case line, openedChannel = <-o.StdErr():
			if !openedChannel {
				if errMsg := common.ReadAll(o.StdErr()); len(errMsg) > 0 {
					err = errors.New(errMsg)
				} else {
					err = io.EOF
				}
			} else {
				foundLine = o.checkLine(line)
			}

		case <-ctx.Done():
			err = errors.New("user cancel")
		}
	}

	return
}
