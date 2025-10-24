package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trip"
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
	err := r.Ch.ExchangeDeclare(
		TripExchange, //name
		"topic",      //type
		true,         //durable
		false,        //auto-deleted
		false,        //internal
		false,        //no-wait
		nil,          //arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %v : %v", TripExchange, err)
	}

	if err := r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated, contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	q, err := r.Ch.QueueDeclare(
		queueName, //name
		true,      //durable
		false,     //delete when unused
		false,     //exclusive
		false,     //no-wait
		nil,       //arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messageTypes {
		err = r.Ch.QueueBind(
			q.Name,   //queue name
			msg,      //routing key
			exchange, //exchange
			false,    //no-wait
			nil,      //arguments
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue to %s : %v", q.Name, err)
		}
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

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	log.Printf("publishing message with routing key: %s", routingKey)

	msgJson, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	return r.Ch.PublishWithContext(ctx,
		TripExchange, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         msgJson,
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
