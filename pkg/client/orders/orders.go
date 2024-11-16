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
	conn, err := grpc.NewClient(
		"localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
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
	_ = userId
	newOrder := &order_proto.Order{
		// UserId:    &order_proto.UUID{Value: userId},
		UserId: &order_proto.UUID{
			Value: "1bf447b8-a129-42a2-b11e-684a801568ff",
		},
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
	fmt.Println("-----order Created ---------")

	fmt.Printf("%+v\n", order.Order.Id)

	// get orders-----
	gOrder, err := c.GetOrder(
		ctx,
		&order_proto.GetOrderRequest{OrderId: order.Order.Id},
	)
	if err != nil {
		return err
	}
	fmt.Println("----- query sucessfull ---------")
	fmt.Printf("%+v\n", gOrder)

	// update product
	upOrder, err := c.UpdateOrder(
		ctx,
		&order_proto.UpdateOrderRequest{
			OrderId:  order.Order.Id,
			Status:   "processing",
			Quantity: 8,
			Total:    8 * product.Price,
		})
	if err != nil {
		return err
	}

	fmt.Println("--------- Order updated ------")
	fmt.Printf("%+v\n", upOrder)

	// get userOrders
	userOrders, err := c.GetUserOrders(
		ctx,
		&order_proto.GetUserOrdersRequest{
			UserId: order.Order.Id,
			Limit:  8,
			Page:   1,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("--------- Getting user orders sucessfull ------")

	for i, uOrder := range userOrders.Orders {
		fmt.Printf("%d ----> %+v\n", i, uOrder)
	}

	gOrders, err := c.GetOrders(
		ctx,
		&order_proto.GetOrdersRequest{Limit: 10, Page: 1},
	)
	if err != nil {
		return err
	}

	fmt.Println("--------- Getting orders sucessfull ------")

	for i, uOrder := range gOrders.Orders {
		fmt.Printf("%d ----> %+v\n", i, uOrder)
	}

	delRes, err := c.DeleteOrder(
		ctx,
		&order_proto.DeleteOrderRequest{OrderId: gOrder.Order.Id},
	)
	if err != nil {
		return err
	}

	fmt.Println("--------- Deleting orders sucessfull ------")
	fmt.Printf("%+v\n", delRes)

	for range 20 {

		order, _ = c.OrderProduct(ctx, &order_proto.OrderProductRequest{
			UserId:    newOrder.UserId,
			ProductId: newOrder.ProductId,
			Quantity:  newOrder.Quantity,
			Total:     newOrder.Total,
		})
		fmt.Println("-----order Created ---------")

		fmt.Printf("%+v\n", order.Order.Id)
	}
	return nil
}
