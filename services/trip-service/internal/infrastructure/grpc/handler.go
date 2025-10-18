package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	pickup := &types.Coordinate{
		Latitude:  req.GetStartLocation().Latitude,
		Longitude: req.GetStartLocation().Longitude,
	}
	destination := &types.Coordinate{
		Latitude:  req.GetEndLocation().Latitude,
		Longitude: req.GetEndLocation().Longitude,
	}

	route, err := h.service.GetRoute(ctx, pickup, destination)
	if err != nil {
		log.Println("error get route: ", err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	estimatedFares := h.service.EstimatePackagesPriceWithRoute(route)
	rideFares, err := h.service.GenerateTripFares(ctx, estimatedFares, req.GetUserID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate trip fares: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFaresProto(rideFares),
	}, nil
}

func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fare := &domain.RideFareModel{
		ID: func() primitive.ObjectID {
			id, err := primitive.ObjectIDFromHex(req.RideFareID)
			if err != nil {
				log.Println(err)
				return primitive.NilObjectID
			}
			return id
		}(),
		UserID: req.UserID,
	}

	trip, err := h.service.CreateTrip(ctx, fare)
	if err != nil {
		log.Println("error create trip: ", err)
		return nil, status.Errorf(codes.Internal, "failed to create trip: %v", err)
	}

	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	}, nil
}
