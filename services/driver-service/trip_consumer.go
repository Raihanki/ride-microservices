package main

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *DriverService
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service *DriverService) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (t *tripConsumer) Listen() error {
	return t.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp091.Delivery) error {
		var tripEvent contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		var payload messaging.TripEventData
		if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
			log.Printf("failed to unmarshal message: %v", err)
			return err
		}

		log.Printf("driver receive message: %+v", payload)

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return t.handleFindAndNotifyDriver(ctx, payload)
		}

		log.Println("unknown trip event")

		return nil
	})
}

func (t *tripConsumer) handleFindAndNotifyDriver(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs := t.service.FindAvailableDrivers(payload.Trip.SelectedRideFare.PackageSlug)

	log.Printf("found suitable drivers: %v", len(suitableIDs))

	if len(suitableIDs) == 0 {
		// notify the rider that no driver is available
		if err := t.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		}); err != nil {
			log.Printf("failed to publish message to exchange: %v", err)
			return err
		}

		return nil
	}

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// notify the rider that a driver is found
	if err := t.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableIDs[0],
		Data:    marshalledEvent,
	}); err != nil {
		log.Printf("failed to publish message to exchange: %v", err)
		return err
	}

	return nil
}
