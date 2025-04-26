package main

import (
	"log"
	"net"

	"github.com/kelcheone/chemistke/cmd/utils"
	orderservice "github.com/kelcheone/chemistke/internal/services/orders"
	order_proto "github.com/kelcheone/chemistke/pkg/grpc/order"
	"google.golang.org/grpc"
)

func main() {
	db, err := utils.GetDB()
	if err != nil {
		log.Panicf("errors connecting to the database: %v", err.Error())
	}

	defer db.Close()
	newOrderService := orderservice.NewOrderService(db)
	grpcServer := grpc.NewServer()

	order_proto.RegisterOrderServiceServer(grpcServer, newOrderService)

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
