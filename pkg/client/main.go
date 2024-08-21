package main

import (
	"log"

	userClient "github.com/kelcheone/chemistke/pkg/client/users"
)

func main() {
	if err := userClient.Init(); err != nil {
		// log.Fatalf("Could not run user client: %v\n", err)
		log.Fatalln(err)
	}
}
