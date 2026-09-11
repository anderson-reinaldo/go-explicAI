package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anderson-reinaldo/go-explicAI/configuration"
)

func main() {
	go configuration.NewApplication().Start()
	shutDown()
}

func shutDown() {
	signalShutdown := make(chan os.Signal, 2)
	signal.Notify(signalShutdown, syscall.SIGINT, syscall.SIGTERM)
	switch <-signalShutdown {
	case syscall.SIGINT:
		fmt.Println("SIGINT signal, explicAI is stopping...")
	case syscall.SIGTERM:
		fmt.Println("SIGTERM signal, explicAI is stopping...")
	}
}
