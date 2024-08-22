package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	productservice "github.com/kelcheone/chemistke/cmd/product-service"
	userservice "github.com/kelcheone/chemistke/cmd/user-service"
	"github.com/kelcheone/chemistke/internal/database"
	"github.com/kelcheone/chemistke/internal/files"
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

	files, err := files.NewFileClient()
	if err != nil {
		log.Fatalf("could not connect to files: %v\n", err)
	}

	newUservice := userservice.NewService(db)
	newProductService := productservice.NewProductService(files, db)

	grpcServer := grpc.NewServer()

	user_proto.RegisterUserServiceServer(grpcServer, newUservice)
	product_proto.RegisterProductServiceServer(grpcServer, newProductService)
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatalf("Could not start the listener: %v\n", err)
	}

	fmt.Println("Serving on port 8090------------------")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Could not serve: %v\n", err)
	}
}
