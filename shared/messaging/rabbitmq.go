package messaging

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
	Ch   *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to rabbitmq: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("fail to create channel rabbitmq: %v", err)
	}

	rmq := &RabbitMQ{conn: conn, Ch: ch}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("fail to setup exchanges and queues rabbitmq: %v", err)
	}

	return rmq, nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	_, err := r.Ch.QueueDeclare(
		"hello", //name
		false,   //durable
		false,   //delete when unused
		false,   //exclusive
		false,   //no-wait
		nil,     //arguments
	)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message string) error {
	return r.Ch.PublishWithContext(ctx,
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return
		}
	}
	if r.Ch != nil {
		if err := r.Ch.Close(); err != nil {
			return
		}
	}
}
