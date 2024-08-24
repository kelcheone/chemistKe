package ordersClient

import (
	"context"
	"fmt"
	"time"

	productsClient "github.com/kelcheone/chemistke/pkg/client/products"
	userClient "github.com/kelcheone/chemistke/pkg/client/users"
	order_proto "github.com/kelcheone/chemistke/pkg/grpc/order"
	product_proto "github.com/kelcheone/chemistke/pkg/grpc/product"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Init() error {
	conn, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	c := order_proto.NewOrderServiceClient(conn)
	proC := product_proto.NewProductServiceClient(conn)
	userC := user_proto.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	// create product and user
	userId, err := userClient.CreateUser(ctx, userC)
	if err != nil {
		return err
	}

	fmt.Println("-----user Creted ---------")
	productId := productsClient.CreateProduct(ctx, proC)
	product, err := productsClient.GetProduct(ctx, proC, productId)
	if err != nil {
		return err
	}
	fmt.Println("----- product Creted ---------")
	// create order
	quantity := 5
	total := product.Price * float32(quantity)
	newOrder := &order_proto.Order{
		UserId:    &order_proto.UUID{Value: userId},
		ProductId: &order_proto.UUID{Value: productId},
		Quantity:  int32(quantity),
		Total:     total,
	}

	order, err := c.OrderProduct(ctx, &order_proto.OrderProductRequest{
		UserId:    newOrder.UserId,
		ProductId: newOrder.ProductId,
		Quantity:  newOrder.Quantity,
		Total:     newOrder.Total,
	})
	fmt.Println("-----order Creted ---------")

	fmt.Printf("%+v\n", order)

	return nil
}
