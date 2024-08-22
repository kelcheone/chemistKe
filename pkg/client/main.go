package main

import (
	"log"

	productsClient "github.com/kelcheone/chemistke/pkg/client/products"
	// userClient "github.com/kelcheone/chemistke/pkg/client/users"
)

func main() {
	// if err := userClient.Init(); err != nil {
	// 	// log.Fatalf("Could not run user client: %v\n", err)
	// 	log.Fatalln(err)
	// }
	if err := productsClient.Init(); err != nil {
		// log.Fatalf("Could not run user client: %v\n", err)
		log.Fatalln(err)
	}
}
