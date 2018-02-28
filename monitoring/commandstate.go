package monitoring

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"
	"bitbucket.org/7phs/tools/common"
)

type commandState struct {
	CoreCmd
	runCount int32
}

func CommandState(parameter Parameter) (*commandState, error) {
	cmdState := &commandState{}
	err := cmdState.Init(parameter)

	return cmdState, err
}

func (o *commandState) Init(parameter Parameter) error {
	_, err := o.init(parameter).prepare()

	return err
}

func (o *commandState) RunCount() int32 {
	return atomic.LoadInt32(&o.runCount)
}

func (o *commandState) Run(ctx context.Context, wait chan<- error) {
	var err error

	err = o.Start()
	if err != nil {
		err = errors.Wrapf(err, "failed to start command for monitoring")
	} else {
		atomic.AddInt32(&o.runCount, 1)

		select {
		case <-common.DeadLineTimer(ctx.Deadline()).C:
			err = errors.New("timeout error")

			// check for std out and unexpected close stdOut and check stdErr
		case _, ok := <-o.StdOut():
			if !ok {
				if errMsg := common.ReadAll(o.StdErr()); len(errMsg) > 0 {
					err = errors.New(errMsg)
				}
			}

		case <-ctx.Done():
			err = errors.New("user cancel")
		}
	}

	wait <- err
}
