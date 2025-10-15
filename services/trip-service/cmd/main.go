package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	h "ride-sharing/services/trip-service/internal/infrastructure/http"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"
	"syscall"
	"time"
)

var (
	httpTripServiceAddr = env.GetString("HTTP_TRIP_SERVICE_ADDR", ":8083")
)

func main() {
	memory := repository.NewMemoryRepository()
	service := service.NewTripServiceImpl(memory)

	mux := http.NewServeMux()

	h := h.HttpHandler{
		Service: service,
	}

	mux.HandleFunc("POST /preview", h.PreviewTrip)

	server := &http.Server{
		Addr:    httpTripServiceAddr,
		Handler: mux,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("server listening on %v", httpTripServiceAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Printf("Error starting server: %v", err)
	case sig := <-shutdown:
		log.Printf("Server is shutting down from %v signal", sig)

		// gracefull shutdown max 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server gracefully %v", err)
			server.Close()
		}
	}
}
