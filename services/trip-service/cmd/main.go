package main

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
)

func main() {
	memory := repository.NewMemoryRepository()
	service := service.NewTripServiceImpl(memory)
	service.CreateTrip(context.Background(), &domain.RideFareModel{UserID: "raihanhori"})
}
