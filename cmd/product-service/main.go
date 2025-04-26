package main

import (
	"log"
	"net"

	"github.com/kelcheone/chemistke/cmd/utils"
	productservice "github.com/kelcheone/chemistke/internal/services/products"
	product_proto "github.com/kelcheone/chemistke/pkg/grpc/product"
	"google.golang.org/grpc"
)

func main() {
	db, err := utils.GetDB()
	if err != nil {
		log.Panicf("errors connecting to the database: %v", err.Error())
	}

	defer db.Close()
	newProductService := productservice.NewProductService(db)
	grpcServer := grpc.NewServer()

	product_proto.RegisterProductServiceServer(grpcServer, newProductService)

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
