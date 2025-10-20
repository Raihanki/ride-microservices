package main

import (
	"context"
	pb "ride-sharing/shared/proto/driver"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer

	service *DriverService
}

func NewGrpcHandler(s *grpc.Server, service *DriverService) *grpcHandler {
	handler := &grpcHandler{
		service: service,
	}

	pb.RegisterDriverServiceServer(s, handler)
	return handler
}

func (h *grpcHandler) RegisterDriver(c context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	driver, err := h.service.RegisterDriver(req.DriverID, req.PackageSlug)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register driver: %v", err)
	}

	resp := &pb.RegisterDriverResponse{
		Driver: driver,
	}
	return resp, nil
}

func (h *grpcHandler) UnregisterDriver(c context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	h.service.UnregisterDriver(req.DriverID)

	return &pb.RegisterDriverResponse{
		Driver: &pb.Driver{Id: req.DriverID},
	}, nil
}
