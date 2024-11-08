package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	cmsservice "github.com/kelcheone/chemistke/cmd/cms-service"
	orderservice "github.com/kelcheone/chemistke/cmd/order-service"
	productservice "github.com/kelcheone/chemistke/cmd/product-service"
	userservice "github.com/kelcheone/chemistke/cmd/user-service"
	"github.com/kelcheone/chemistke/internal/database"
	cms_proto "github.com/kelcheone/chemistke/pkg/grpc/cms"
	order_proto "github.com/kelcheone/chemistke/pkg/grpc/order"
	product_proto "github.com/kelcheone/chemistke/pkg/grpc/product"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Could not load env variables: %v\n", err)
	}
	connStr := fmt.Sprintf(

		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",

		os.Getenv("DB_HOST"),

		os.Getenv("DB_PORT"),

		os.Getenv("DB_USER"),

		os.Getenv("DB_PASSWORD"),

		os.Getenv("DB_NAME"),
	)
	db, err := database.NewDatabase("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v\n", err)
	}

	defer db.Close()

	newUservice := userservice.NewService(db)
	newProductService := productservice.NewProductService(db)
	newOrderService := orderservice.NewOrderService(db)
	newCmsService := cmsservice.NewCmsService(db)

	grpcServer := grpc.NewServer()

	user_proto.RegisterUserServiceServer(grpcServer, newUservice)
	product_proto.RegisterProductServiceServer(grpcServer, newProductService)
	order_proto.RegisterOrderServiceServer(grpcServer, newOrderService)
	cms_proto.RegisterCmsServiceServer(grpcServer, newCmsService)
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatalf("Could not start the listener: %v\n", err)
	}

	fmt.Println("Serving on port 8090------------------")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Could not serve: %v\n", err)
	}
}
