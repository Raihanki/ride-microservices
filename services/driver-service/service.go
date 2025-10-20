package main

import (
	math "math/rand/v2"
	pb "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/util"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type DriverService struct {
	drivers []*driverInMap
	mu      sync.RWMutex
}

type driverInMap struct {
	Driver *pb.Driver
	// TODO: route
}

func NewDriverService() *DriverService {
	return &DriverService{
		drivers: make([]*driverInMap, 0),
	}
}

func (s *DriverService) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	randomIndex := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIndex]

	// we can ignore this property for now, but it must be sent to the frontend.
	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver := &pb.Driver{
		Geohash:        geohash,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:           "Lando Norris",
		Id:             driverId,
		ProfilePicture: util.GetRandomAvatar(randomIndex),
		CarPlate:       GenerateRandomPlate(),
		PackageSlug:    packageSlug,
	}

	// Add driver to list
	s.drivers = append(s.drivers, &driverInMap{Driver: driver})
	return driver, nil
}

func (s *DriverService) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, d := range s.drivers {
		if d.Driver.Id == driverId {
			// delete driver from list
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
		}
	}
}
