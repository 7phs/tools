package shutdown

import (
	"os"
	"os/signal"
)

var gateKeeper struct {
	interrupt chan os.Signal
	shutdown  chan bool
}

func Register(callback func()) {
	go func() {
		<-gateKeeper.shutdown
		callback()
	}()
}

func Shutdown() {
	close(gateKeeper.interrupt)
}

func init() {
	gateKeeper.interrupt = make(chan os.Signal, 2)
	gateKeeper.shutdown = make(chan bool)

	signal.Notify(gateKeeper.interrupt, os.Interrupt) // CTRL-C
	go func() {
		<-gateKeeper.interrupt
		close(gateKeeper.shutdown)
	}()
}
