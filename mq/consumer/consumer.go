package consumer

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"log"
	"queuev2/api"
	"queuev2/httpclient"
	"time"
)

type MQConsumer struct {
	amqpURL    string
	exchange   string
	tag        string
	queueName  string
	bindingKey string
	conn       *amqp.Connection
}

func NewMQConsumer(amqpURI, exchange, tag, queueName, bindingKey string) *MQConsumer {
	c := &MQConsumer{
		amqpURL:    amqpURI,
		exchange:   exchange,
		tag:        tag,
		queueName:  queueName,
		bindingKey: bindingKey,
	}

	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Fatalf("error:: amqp dial: %+v", err)
	}

	c.conn = connection
	return c
}

func (c *MQConsumer) Start() {
	channel, err := c.conn.Channel()
	if err != nil {
		log.Fatalf("error:: getting channel: %+v \n", err)
	}

	if err = channel.QueueBind(
		c.queueName,  // name of the queue
		c.bindingKey, // bindingKey
		c.exchange,   // sourceExchange
		false,        // noWait
		nil,          // arguments
	); err != nil {
		log.Fatalf("error:: creating binding: %+v \n", err)
	}

	// set prefect = 1
	err = channel.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("error:: set basic.qos: %v", err)
	}

	deliveries, err := channel.Consume(
		c.queueName, // name
		c.tag,       // consumerTag,
		false,       // noAck
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("error:: consumer start: %v", err)
	}

	go c.handleMessages(deliveries)
}

func (c *MQConsumer) handleMessages(deliveries <-chan amqp.Delivery) {
	agentURL := "sip:1111@freeswitch-registrar-10x.i3clogic.com:5508"
	for d := range deliveries {
		log.Printf(
			"received task at consumer: [%v] %+v",
			d.DeliveryTag,
			string(d.Body),
		)
		task := &api.Task{}
		err := json.Unmarshal(d.Body, task)
		if err != nil {
			log.Println("error:: ", err)
		}

		callUUID := task.CallData["call_uuid"]
		log.Println("debug: finding agent for call_uuid", callUUID)
		time.Sleep(10 * time.Second)

		log.Printf("debug: agent %s found for call_uuid %s", agentURL, callUUID)

		// Execute modify on call
		err = c.transferToAgent(agentURL, callUUID)
		if err != nil {
			log.Println("error:: ", err)
		}
		d.Ack(false)
	}
}

func (c *MQConsumer) transferToAgent(agentURL, callUUID string) error {
	url := "http://52.71.132.13:8888/v1.0/accounts/123/calls/" + callUUID + "/modify"
	body := map[string]string{"cccml": "<Response><Say>Modify successfull</Say><Dial><Sip>" + agentURL + "</Sip></Dial></Response>"}
	_, err := httpclient.Post(body, url, map[string]string{contentType: contentTypeJSON})
	if err != nil {
		return err
	}
	return nil
}
