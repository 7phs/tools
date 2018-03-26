package monitoring

import (
	"io"
	"os/exec"
	"sync"

	"github.com/pkg/errors"

	"github.com/7phs/tools/common"
)

type Parameter interface {
	WorkDir() string
	Command() string
	ToArgs() []string
}

type CoreCmd struct {
	cmd *exec.Cmd

	stdoutStream chan string
	stderrStream chan string
	wait         chan bool

	monitoringCh    chan bool
	monitoringClose func()
	monitoringWait  sync.WaitGroup
}

func NewCoreCmd(parameter Parameter) (*CoreCmd, error) {
	return (&CoreCmd{}).
		init(parameter).
		prepare()
}

func (o *CoreCmd) init(parameter Parameter) *CoreCmd {
	o.cmd =  exec.Command(parameter.Command(), parameter.ToArgs()...)
	o.cmd.Dir = parameter.WorkDir()

	return o
}

func (o *CoreCmd) prepare() (*CoreCmd, error) {
	var (
		errOut error
		errErr error
	)

	o.monitoringClose = func() {}

	o.stdoutStream, errOut = o.reader2Stream(o.cmd.StdoutPipe())
	o.stderrStream, errErr = o.reader2Stream(o.cmd.StderrPipe())

	return o, common.SeveralErrors("failed to prepare streams",
		errors.Wrap(errOut, "failed to prepare stdoutStream"),
		errors.Wrap(errErr, "failed to prepare stderrStream"),
	)
}

func (o *CoreCmd) reader2Stream(reader io.Reader, err error) (chan string, error) {
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get pipe")
	}

	if reader == nil {
		return nil, errors.New("reader is nil")
	}

	return common.MakeStream(reader), nil
}

func (o *CoreCmd) StdOut() <-chan string {
	return o.stdoutStream
}

func (o *CoreCmd) StdErr() <-chan string {
	return o.stderrStream
}

func (o *CoreCmd) Wait() <-chan bool {
	return o.wait
}

func (o *CoreCmd) Start() error {
	if err := o.cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start command")
	}

	o.wait = make(chan bool)
	go func() {
		o.cmd.Wait()
		close(o.wait)
	}()

	return nil
}

func (o *CoreCmd) Kill() error {
	if err:=o.checkProcessState(); err!=nil {
		return err
	}

	o.monitoringClose()
	o.monitoringWait.Wait()

	err := o.cmd.Process.Kill()

	return errors.Wrap(err, "failed to kill the command")
}

func (o *CoreCmd) monitoring() <-chan bool {
	return o.monitoringCh
}

func (o *CoreCmd) Monitoring(done DoneFunc) {
	if err:=o.checkProcessState(); err!=nil {
		done(err)
		return
	}

	o.monitoringCh = make(chan bool)
	o.monitoringClose = func() {
		close(o.monitoringCh)
	}
	o.monitoringWait.Add(1)

	go func() {
		select {
		case <-o.Wait():
			done(nil)

		case <-o.monitoring():
		}

		o.monitoringWait.Done()
	}()
}

func (o *CoreCmd) IsExited() bool {
	return o.cmd!=nil && o.cmd.ProcessState!=nil && o.cmd.ProcessState.Exited()
}

func (o *CoreCmd) checkProcessState() error {
	// check start
	if o.cmd == nil || o.cmd.Process == nil {
		return errors.New("failed to kill process id-less")
	}
	// check finish
	if o.IsExited() {
		return errors.New("already killed")
	}

	return nil
}
