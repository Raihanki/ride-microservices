package main

import (
	"context"
	"log"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbitmq,
	}
}

func (t *tripConsumer) Listen() error {
	return t.rabbitmq.ConsumeMessages("hello", func(ctx context.Context, msg amqp091.Delivery) error {
		log.Printf("driver receive message: %v", msg)
		return nil
	})
}
