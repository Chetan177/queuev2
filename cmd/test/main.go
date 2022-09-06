package main

import (
	"log"
	"net/http"

	rabbit "github.com/michaelklishin/rabbit-hole/v2"
)

func main() {
	transport := &http.Transport{}
	rmq, _ := rabbit.NewTLSClient("https://b-d5f9cd82-8b67-46e1-a98b-8706a2766797.mq.us-east-1.amazonaws.com", "voice", "voice@3clogic", transport)
	resp, err := rmq.Overview()
	if err != nil {
		log.Fatalln("Error: connection", err)
	}
	log.Println(resp)

	nodes, err := rmq.ListExchanges()
	if err != nil {
		log.Fatalln("Error: connection", err)
	}
	log.Printf("%+v", nodes)
}
