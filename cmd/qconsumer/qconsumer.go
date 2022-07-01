package main

import (
	"log"
	"os"
	"os/signal"
	"queuev2/mq/consumer"
	"syscall"
)

const (
	exchange   = "queuev2-exchange"
	queueName  = "QID_MTIzX3NhbGVzX3F1ZXVlNQ=="
	tag        = queueName + "_ConsumerTag"
	bindingKey = queueName + "_rKey"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	c := consumer.NewMQConsumer("amqp://guest:guest@localhost:5672/",
		exchange,
		tag,
		queueName,
		bindingKey)

	c.Start()

	log.Println("consumer started successfully")
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	<-done

}
