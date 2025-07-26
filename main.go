package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hemupadhyay26/ssh-cred-manager-cli/cmd"
)

// version is set at build time using -ldflags
var version = "dev"

func main() {
	// Handle interrupt signal (Ctrl+C)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("bye")
		os.Exit(0)
	}()

	if err := cmd.Execute(version); err != nil {
		// If error is due to interrupt, do nothing (handled above)
		// Otherwise, print error
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
