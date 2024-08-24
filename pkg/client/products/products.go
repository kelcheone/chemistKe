package productsClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	product_proto "github.com/kelcheone/chemistke/pkg/grpc/product"
)

func Init() error {
	conn, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		Category:    gofakeit.ProductCategory(),
		SubCategory: gofakeit.Product().Categories[0],
		Brand:       gofakeit.Company(),
		Price:       float32(gofakeit.Product().Price),
		Quantity:    int32(gofakeit.Number(0, 100)),
	}

	createRes := CreateProduct(ctx, c)
	// if err != nil {
	// 	log.Fatalf("Could not create product %v\n", err)
	// }

	fmt.Printf("%+v\n", createRes)

	// ------------- GET PRODUCT -------------
	getRes, err := c.GetProduct(ctx, &product_proto.GetProductRequest{Id: &product_proto.UUID{Value: createRes}})
	if err != nil {
		log.Fatalf("Could not get product %v\n", err)
	}

	fmt.Printf("%+v\n", getRes)

	// ------------- UPDATE PRODUCT -------------
	product.Id = &product_proto.UUID{Value: createRes}
	product.Name = gofakeit.Product().Name
	product.Description = gofakeit.Product().Description
	product.Category = gofakeit.ProductCategory()
	product.SubCategory = gofakeit.Product().Categories[0]
	product.Brand = gofakeit.Company()
	product.Price = float32(gofakeit.Product().Price)
	product.Quantity = int32(gofakeit.Number(0, 100))

	updateRes, err := c.UpdateProduct(ctx, &product_proto.UpdateProductRequest{Product: product})
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

	url, err := c.GetUploadURL(ctx, &product_proto.GetUploadURLRequest{Id: &product_proto.UUID{Value: createRes}, FileName: "image1.png"})
	fmt.Printf("%+v\n", url)

	imagePath := "./public/images/image1.png"
	err = uploadFile(imagePath, url.Url)
	if err != nil {
		return err
	}
	sT := time.Now()
	images, err := c.GetProductImages(ctx, &product_proto.GetProductImagesRequest{ProductId: &product_proto.UUID{Value: createRes}})
	if err != nil {
		log.Fatalf("Could not get images: %v\n", err)
	}

	for i, image := range images.Urls {
		fmt.Printf("%d ----------> %s\n", i, image)
	}
	fmt.Println(time.Since(sT))

	return nil
}

func uploadFile(filePath string, url string) error {
	file, err := os.Open(filePath)
	defer file.Close()
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPut, url, buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "multipart/form-data")
	// 'x-amz-acl': 'public-read' -- This header is required for public read ACL for Digital ocean
	request.Header.Set("x-amz-acl", "public-read")
	client := &http.Client{}
	resp, err := client.Do(request)

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Upload successful")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Upload failed. Status: %s, Body: %s\n", resp.Status, string(body))
	}
	return err
}

func CreateProduct(ctx context.Context, c product_proto.ProductServiceClient) string {
	product := &product_proto.Product{
		Name:        gofakeit.Product().Name,
		Description: gofakeit.Product().Description,
		Category:    gofakeit.ProductCategory(),
		SubCategory: gofakeit.Product().Categories[0],
		Brand:       gofakeit.Company(),
		Price:       float32(gofakeit.Product().Price),
		Quantity:    int32(gofakeit.Number(0, 100)),
	}

	createRes, err := c.CreateProduct(ctx, &product_proto.CreateProductRequest{Product: product})
	if err != nil {
		log.Fatalf("Could not create product %v\n", err)
	}

	return createRes.Id.Value
}

func GetProduct(ctx context.Context, c product_proto.ProductServiceClient, id string) (*product_proto.Product, error) {
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
