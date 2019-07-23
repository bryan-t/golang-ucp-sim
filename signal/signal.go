// Package signal is a convenience package for handling operating system signals.
// Currently, only SIGTERM and SIGINT will trigger and exit to the Listen func.
package signal

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// REVIEW: For simplicity, I've implemented this as a singleton, it would be more
// proper to put this into a struct.
var (
	handlers = make(map[os.Signal]func())
	stop     = make(chan int, 1)
	done     = make(chan int, 1)
)

// Handle adds a handler for a signal. If a handler is
// already registered for a signal, it will be overwritten.
func Handle(sig os.Signal, handler func()) {
	handlers[sig] = handler
}

// Listen starts listening for signals and calls
// the corresponding handler.
func Listen() {

	defer func() {
		log.Println("Listen done\n")
		close(done)
	}()

	log.Printf("Listen start\n")

	notify := make(chan os.Signal, 1)

	for sig, _ := range handlers {
		signal.Notify(notify, sig)
	}

	for {
		select {
		case sig := <-notify:

			log.Printf("Received %s\n", sig)
			handlers[sig]()

			if sig == syscall.SIGTERM ||
				sig == syscall.SIGINT {
				return
			}

		case <-stop:
			return
		}
	}

}

// Stop gracefully stops Listen.
func Stop() {
	log.Printf("Stopping Listen()\n")
	stop <- 1 // tell the listener to stop
	log.Printf("Waiting for Listen() to stop\n")
	<-done // wait for listener to stop
}

// Wait will block until Listen finishes.
func Wait() {
	log.Printf("Waiting for Listen() to stop\n")
	<-done // wait for listener to stop
}
