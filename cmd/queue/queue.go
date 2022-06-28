package main

import (
	"os"
	"os/signal"
	"syscall"

	"queuev2/api"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	server := api.GetServerInstance(9898, nil)
	server.StartServer()

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		done <- true
	}()

	<-done

}
