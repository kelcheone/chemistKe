package productsClient

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	product_proto "github.com/kelcheone/chemistke/pkg/grpc/product"
)

func Init() error {
	conn, err := grpc.NewClient(
		"localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	// fmt.Printf("%+v\n", conn)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := product_proto.NewProductServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)

	defer cancel()

	//------------- ADD PRODUCT -------------
	product := &product_proto.Product{
		Name:        gofakeit.Product().Name,
		Description: gofakeit.Product().Description,
		// Category:    gofakeit.ProductCategory(),
		// SubCategory: gofakeit.Product().Categories[0],
		// Brand:       gofakeit.Company(),
		Price:    float32(gofakeit.Product().Price),
		Quantity: int32(gofakeit.Number(0, 100)),
	}

	createRes := CreateProduct(ctx, c)
	// if err != nil {
	// 	log.Fatalf("Could not create product %v\n", err)
	// }

	fmt.Printf("%+v\n", createRes)

	// ------------- GET PRODUCT -------------
	getRes, err := c.GetProduct(
		ctx,
		&product_proto.GetProductRequest{
			Id: &product_proto.UUID{Value: createRes},
		},
	)
	if err != nil {
		log.Fatalf("Could not get product %v\n", err)
	}

	fmt.Printf("%+v\n", getRes)

	// ------------- UPDATE PRODUCT -------------
	product.Id = &product_proto.UUID{Value: createRes}
	product.Name = gofakeit.Product().Name
	product.Description = gofakeit.Product().Description
	// product.CategoryId = gofakeit.ProductCategory()
	// product.SubCategoryId = gofakeit.Product().Categories[0]
	// product.BrandId = gofakeit.Company()
	product.Price = float32(gofakeit.Product().Price)
	product.Quantity = int32(gofakeit.Number(0, 100))

	updateRes, err := c.UpdateProduct(
		ctx,
		&product_proto.UpdateProductRequest{Product: product},
	)
	if err != nil {
		log.Fatalf("Could not update product %v\n", err)
	}

	fmt.Printf("%+v\n", updateRes)

	// ------------- GET ALL PRODUCTS -------------
	paginatedReq := &product_proto.GetProductsRequest{
		Limit: 10,
		Page:  1,
	}

	getAllRes, err := c.GetProducts(ctx, paginatedReq)
	if err != nil {
		log.Fatalf("Could not get all products %v\n", err)
	}
	for i, p := range getAllRes.Products {
		fmt.Printf("%d------> %s, \t kes%.2f\n", i, p.Name, p.Price)
	}

	// for range 20 {
	// 	_ = CreateProduct(ctx, c)
	// }

	return nil
}

func CreateProduct(
	ctx context.Context,
	c product_proto.ProductServiceClient,
) string {
	product := &product_proto.Product{
		Name:        gofakeit.Product().Name,
		Description: gofakeit.Product().Description,
		// CategoryId:    gofakeit.ProductCategory(),
		// SubCategoryId: gofakeit.Product().Categories[0],
		// BrandId:       gofakeit.Company(),
		Price:    float32(gofakeit.Product().Price),
		Quantity: int32(gofakeit.Number(0, 100)),
	}

	createRes, err := c.CreateProduct(
		ctx,
		&product_proto.CreateProductRequest{Product: product},
	)
	if err != nil {
		log.Fatalf("Could not create product %v\n", err)
	}

	return createRes.Id.Value
}

func GetProduct(
	ctx context.Context,
	c product_proto.ProductServiceClient,
	id string,
) (*product_proto.Product, error) {
	getRes, err := c.GetProduct(
		ctx,
		&product_proto.GetProductRequest{
			Id: &product_proto.UUID{
				Value: id,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return getRes.Product, nil
}
