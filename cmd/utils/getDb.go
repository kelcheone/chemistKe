package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelcheone/chemistke/internal/database"
)

func GetDB() (database.DB, error) {
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
	log.Printf("Connecting to database with connection string: %s", connStr)
	db, err := database.NewDatabase("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v\n", err)
	}

	return db, nil
}
