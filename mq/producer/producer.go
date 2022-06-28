package producer

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type MQProducer struct {
	amqpURL      string
	exchange     string
	exchangeType string
	conn         *amqp.Connection
}

func NewMQProducer(amqpURI, exchange, exchangeType string) *MQProducer {
	p := &MQProducer{
		amqpURL:      amqpURI,
		exchange:     exchange,
		exchangeType: exchangeType,
	}

	log.Printf("dialing %q", amqpURI)
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Fatalf("error:: amqp dial: %+v", err)
	}

	p.conn = connection
	return p
}

func (p *MQProducer) Start() {
	if err := p.declareExchange(); err != nil {
		log.Fatalf("error:: prodcuer start: %+v", err)
	}
}

func (p *MQProducer) declareExchange() error {
	channel, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("error:: getting channel: %+v", err)
	}
	log.Printf("got Channel, declaring %q Exchange (%q)", p.exchangeType, p.exchange)
	if err := channel.ExchangeDeclare(
		p.exchange,     // name
		p.exchangeType, // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // noWait
		nil,            // arguments
	); err != nil {
		return fmt.Errorf("error:: exchange declare: %+v", err)
	}

	channel.Close()

	return nil
}

func (p *MQProducer) PublishMessage(routingKey string, body []byte, priority uint8) error {
	channel, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("error:: getting channel: %+v", err)
	}
	if err = channel.Publish(
		p.exchange, // publish to an exchange
		routingKey, // routing to 0 or more queues
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        priority,       // 0-9
		},
	); err != nil {
		return fmt.Errorf("error:: exchange Publish: %+v", err)
	}
	channel.Close()
	return nil
}
