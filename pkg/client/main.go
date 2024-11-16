package main

import (
	"log"

	// cmsClient "github.com/kelcheone/chemistke/pkg/client/cms"
	ordersClient "github.com/kelcheone/chemistke/pkg/client/orders"
	// productsClient "github.com/kelcheone/chemistke/pkg/client/products"
	// userClient "github.com/kelcheone/chemistke/pkg/client/users"
)

func main() {
	// if err := userClient.Init(); err != nil {
	// 	// log.Fatalf("Could not run user client: %v\n", err)
	// 	log.Fatalln(err)
	// }
	// if err := productsClient.Init(); err != nil {
	// 	log.Fatalln(err)
	// }

	if err := ordersClient.Init(); err != nil {
		log.Fatalln(err)
	}
	// if err := cmsClient.Init(); err != nil {
	// 	log.Fatalln(err)
	// }
}
