package service

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripServiceImpl struct {
	repository domain.TripRepository
}

func NewTripServiceImpl(repository domain.TripRepository) *TripServiceImpl {
	return &TripServiceImpl{
		repository: repository,
	}
}

func (s *TripServiceImpl) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	trip := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
	}

	newTrip, err := s.repository.CreateTrip(ctx, trip)
	if err != nil {
		log.Println(err)
	}

	return newTrip, nil
}
