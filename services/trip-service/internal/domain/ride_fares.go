package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	pb "ride-sharing/shared/proto/trip"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string // ex: "luxury"
	TotalPriceInCents float64
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	var result []*pb.RideFare
	for _, f := range fares {
		result = append(result, f.ToProto())
	}
	return result
}
