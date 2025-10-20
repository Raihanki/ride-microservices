package main

import (
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"
	pb "ride-sharing/shared/proto/driver"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleDriverWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("userID is required")
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")
	if userID == "" {
		log.Println("packageSlug is required")
		return
	}

	c, err := grpc_clients.NewDriverServiceClient()
	if err != nil {
		log.Fatal(err)
	}

	ctx := r.Context()

	defer func() {
		c.Client.UnregisterDriver(ctx, &pb.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		})
		c.Close()
		log.Printf("Driver Unregisterd ID: %v", userID)
	}()

	driver, err := c.Client.RegisterDriver(
		ctx,
		&pb.RegisterDriverRequest{
			DriverID:    userID,
			PackageSlug: packageSlug,
		},
	)
	if err != nil {
		log.Printf("failed to register driver: %v", err)
		http.Error(w, "failed to register driver", http.StatusInternalServerError)
		return
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driver.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("error sending message in websocket: %v", err)
		return
	}
}

func handleRiderWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userID := r.URL.Query().Get("userID")
	if userID == "" {
		log.Println("userID is required")
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading message in websocket: %v", err)
			break
		}

		log.Printf("received message: %s", message)
	}
}
