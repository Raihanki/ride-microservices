package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	GrpcAddr = ":9092"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, syscall.SIGTERM)
		<-signalCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	service := NewDriverService()

	// RabbitMQ setup
	rabbitmq, err := messaging.NewRabbitMQ(
		env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	log.Println("starting rabbitmq connection")

	// starting the grpc server
	grpcserver := grpcserver.NewServer()
	NewGrpcHandler(grpcserver, service)

	// rabbitmq listener
	consumer := NewTripConsumer(rabbitmq)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
	}()

	log.Printf("starting grpc server Driver Service on port %s", lis.Addr().String())

	go func() {
		if err := grpcserver.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	// wait for the shutdown signal
	<-ctx.Done()
	log.Println("shutting down gRPC server ...")
	grpcserver.GracefulStop()
}
