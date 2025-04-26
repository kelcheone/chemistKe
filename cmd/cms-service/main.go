package main

import (
	"log"
	"net"

	"github.com/kelcheone/chemistke/cmd/utils"
	cmsservice "github.com/kelcheone/chemistke/internal/services/cms"
	cms_proto "github.com/kelcheone/chemistke/pkg/grpc/cms"
	"google.golang.org/grpc"
)

func main() {
	db, err := utils.GetDB()
	if err != nil {
		log.Panicf("errors connecting to the database: %v", err.Error())
	}

	defer db.Close()
	newCmsService := cmsservice.NewCmsService(db)
	grpcServer := grpc.NewServer()

	cms_proto.RegisterCmsServiceServer(grpcServer, newCmsService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
