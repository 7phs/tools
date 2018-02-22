package monitoring

import (
	"context"
	"fmt"
	"sync"
	"github.com/pkg/errors"
)

// closedChan is a reusable closed channel.
var closedChan = make(chan struct{})

func init() {
	close(closedChan)
}

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc does not wait for the work to stop.
// After the first call, subsequent calls to a CancelFunc do nothing.
type DoneFunc func(err error)

// A valueCtx carries a key-value pair. It implements Value for that key and
// delegates all other calls to the embedded Context.
type commandCtx struct {
	context.Context
	sync.Mutex

	command monitoringCommand
	err     error
	wait    chan struct{}
}

func CommandCtx(parent context.Context, command monitoringCommand) (context.Context, DoneFunc) {
	ctx := &commandCtx{
		Context: parent,
		command: command,
	}

	return ctx, func(err error) {
		fmt.Println("CommandCtx:DoneFunc")

		fmt.Println("CommandCtx:DoneFunc:ctx.done(err)")
		ctx.done(err)

		if parentCtx, ok := parent.(*commandCtx); ok {
			fmt.Println("StageCtx:DoneFunc:parentCtx.done(err)")
			parentCtx.done(err)
		}
	}
}

func (o *commandCtx) String() string {
	return fmt.Sprintf("%v.CommandCtx(%v)", o.Context, o.command)
}

func (o *commandCtx) Value(key interface{}) interface{} {
	if key == "command" {
		return o.command
	}
	return o.Context.Value(key)
}

func (o *commandCtx) Done() <-chan struct{} {
	o.Lock()
	if o.wait == nil {
		o.wait = make(chan struct{})
	}

	w := o.wait

	go func() {
		<-o.Context.Done()

		o.done(errors.New("user cancel"))
	}()

	o.Unlock()

	return w
}

func (o *commandCtx) HasError() error {
	return o.err
}

func (o *commandCtx) done(v interface{}) {
	fmt.Println("commandCtx:done(", v, ")")

	o.Lock()

	// already closed
	if o.wait == closedChan {
		o.Unlock()
		return
	}

	o.err, _ = v.(error)

	if o.wait != nil {
		fmt.Println("commandCtx:done:close(", o.wait, ")")
		close(o.wait)
	}
	o.wait = closedChan

	o.Unlock()
}
