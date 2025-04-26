package main

import (
	"log"
	"net"

	"github.com/kelcheone/chemistke/cmd/utils"
	userservice "github.com/kelcheone/chemistke/internal/services/users"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"google.golang.org/grpc"
)

func main() {
	db, err := utils.GetDB()
	if err != nil {
		log.Panicf("errors connecting to the database: %v", err.Error())
	}

	defer db.Close()
	newUserService := userservice.NewService(db)
	grpcServer := grpc.NewServer()

	user_proto.RegisterUserServiceServer(grpcServer, newUserService)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
