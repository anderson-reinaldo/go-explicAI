package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/anderson-reinaldo/go-explicAI/configuration"
	"github.com/anderson-reinaldo/go-explicAI/internal/infrastructure/log"
)

func main() {
	c := configuration.Init()
	go configuration.NewApplication(c).Start()
	shutDown()
}

func shutDown() {
	signalShutdown := make(chan os.Signal, 2)
	signal.Notify(signalShutdown, syscall.SIGINT, syscall.SIGTERM)
	switch <-signalShutdown {
	case syscall.SIGINT:
		log.LogInfo(context.Background(), "SIGINT signal, explicAI is stopping...")
	case syscall.SIGTERM:
		log.LogInfo(context.Background(), "SIGTERM signal, explicAI is stopping...")
	}
}
