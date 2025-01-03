package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s := InitSensitiveCrawler()
	go func() {
		s.Run(ctx)
	}()
	<-quit
}
