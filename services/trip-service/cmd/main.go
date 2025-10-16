package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	GrpcAddr = ":9093"
)

func main() {
	memory := repository.NewMemoryRepository()
	service.NewTripServiceImpl(memory)

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

	grpcserver := grpcserver.NewServer()

	log.Printf("starting grpc server Trip Service on port %s", lis.Addr().String())

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
