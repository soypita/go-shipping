package main

import (
	"context"
	"log"

	pb "github.com/soypita/go-shipping/proto/consignment"
	vesselProto "github.com/soypita/shippy-service-vessel/proto/vessel"
)

type handler struct {
	repo         repository
	vesselClient vesselProto.VesselServiceClient
}

func (s *handler) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {

	vesselResp, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})
	log.Printf("Found vessel: %s \n", vesselResp.Vessel.Name)
	if err != nil {
		return err
	}

	req.VesselId = vesselResp.Vessel.Id

	if err = s.repo.Create(ctx, MarshalConsignment(req)); err != nil {
		return err
	}

	res.Created = true
	res.Consignment = req
	return nil
}

func (s *handler) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	consignments, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil
	}
	res.Consignments = UnmarshalConsignmentCollection(consignments)
	return nil
}
