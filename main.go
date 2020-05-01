package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
	pb "github.com/soypita/go-shipping/proto/consignment"
	userService "github.com/soypita/shippy-service-user/proto/user"
	vesselProto "github.com/soypita/shippy-service-vessel/proto/vessel"
)

const (
	defaultHost = "datastore:27017"
)

func AuthWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, resp interface{}) error {
		meta, ok := metadata.FromContext(ctx)
		if !ok {
			return errors.New("no auth meta-data found in request")
		}

		token, ok := meta["Token"]
		if !ok {
			return errors.New("no token in meta")
		}

		log.Println("Authenticating with token: ", token)

		// Auth here
		authClient := userService.NewUserServiceClient("go.micro.srv.user", client.DefaultClient)
		_, err := authClient.ValidateToken(context.Background(), &userService.Token{
			Token: token,
		})
		if err != nil {
			return err
		}
		err = fn(ctx, req, resp)
		return err
	}
}

func main() {
	// Create a new service. Optionally include some options here.
	srv := micro.NewService(
		micro.Name("shippy.consignment.service"),
		micro.WrapHandler(AuthWrapper),
	)

	// Init will parse the command line flags.
	srv.Init()

	uri := os.Getenv("DB_HOST")
	if uri == "" {
		uri = defaultHost
	}

	client, err := CreateClient(context.Background(), uri, 0)
	if err != nil {
		log.Panic(err)
	}
	defer client.Disconnect(context.Background())

	consignmentCollection := client.Database("shippy").Collection("consignments")

	repository := &MongoRepository{consignmentCollection}
	vesselClient := vesselProto.NewVesselServiceClient("shippy.service.vessel", srv.Client())

	// Register handler
	h := &handler{repository, vesselClient}

	pb.RegisterShippingServiceHandler(srv.Server(), h)

	// Run the server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
