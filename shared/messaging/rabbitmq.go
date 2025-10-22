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

type MessageHandler func(context.Context, amqp.Delivery) error

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
		true,    //durable
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

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	// fair dispatch
	err := r.Ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set Qos: %v", err)
	}

	msgs, err := r.Ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		for msg := range msgs {
			log.Printf("received a message %v", msg)
			if err := handler(ctx, msg); err != nil {
				log.Printf("failed to handle the message: %v", err)

				if err := msg.Nack(false, false); err != nil {
					log.Printf("ERROR: failed to nack message %v", err)
				}

				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("ERROR: failed to ack message %v", err)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message string) error {
	return r.Ch.PublishWithContext(ctx,
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(message),
			DeliveryMode: amqp.Persistent,
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
