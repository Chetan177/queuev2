package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"queuev2/api"
	"queuev2/mq/producer"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	mq := producer.NewMQProducer("amqp://guest:guest@localhost:5672/", "queuev2-exchange", "direct")
	mq.Start()
	server := api.GetServerInstance(9898, mq)
	err := server.StartServer()
	if err != nil {
		log.Fatalln(err)
	}

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		done <- true
	}()

	<-done

}
