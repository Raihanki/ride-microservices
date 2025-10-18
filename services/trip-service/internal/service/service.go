package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	"ride-sharing/shared/types"

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

func (s *TripServiceImpl) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OSRMAPIResponse, error) {
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude, pickup.Latitude,
		destination.Longitude, destination.Latitude,
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to read route from ORSM API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response: %v", err)
	}

	var routeResponse tripTypes.OSRMAPIResponse
	if err := json.Unmarshal(body, &routeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the response: %v", err)
	}

	return &routeResponse, nil
}

func (s *TripServiceImpl) EstimatePackagesPriceWithRoute(route *tripTypes.OSRMAPIResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()
	estimatedPrice := make([]*domain.RideFareModel, len(baseFares))

	for i, p := range baseFares {
		estimatedPrice[i] = estimateFare(route, p)
	}

	return estimatedPrice
}

func (s *TripServiceImpl) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFareModel, userId string) ([]*domain.RideFareModel, error) {
	fares := make([]*domain.RideFareModel, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()
		fare := &domain.RideFareModel{
			UserID:            userId,
			ID:                id,
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug:       f.PackageSlug,
		}

		if err := s.repository.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save trip fare: %v", err)
		}

		fares[i] = fare
	}

	return fares, nil
}

func estimateFare(route *tripTypes.OSRMAPIResponse, fare *domain.RideFareModel) *domain.RideFareModel {
	pricing := tripTypes.DefaultPricingConfig()
	carPackagePrice := fare.TotalPriceInCents

	distanceKM := route.Routes[0].Distance
	durationInMinute := route.Routes[0].Duration

	distanceFare := distanceKM * pricing.PricePerUnitOfDistance
	timeFare := durationInMinute * pricing.PricePerMinute
	totalPrice := carPackagePrice + distanceFare + timeFare

	return &domain.RideFareModel{
		PackageSlug:       fare.PackageSlug,
		TotalPriceInCents: totalPrice,
	}
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}
