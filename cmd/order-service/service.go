package orderservice

import (
	"context"
	"fmt"

	"github.com/kelcheone/chemistke/internal/database"
	pb "github.com/kelcheone/chemistke/pkg/grpc/order"
)

type OrderService struct {
	db database.DB
	pb.UnimplementedOrderServiceServer
}

func NewOrderService(db database.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) OrderProduct(ctx context.Context, req *pb.OrderProductRequest) (*pb.OrderProductResponse, error) {
	stmt := `INSERT INTO orders (user_id, product_id, quantity, total) VALUES ($1, $2, $3, $4) RETURNING id`

	var orderId string
	err := s.db.QueryRow(stmt, req.UserId.Value, req.ProductId.Value, req.Quantity, req.Total).Scan(&orderId)
	if err != nil {
		return nil, err
	}
	order := &pb.Order{
		Id:        &pb.UUID{Value: orderId},
		UserId:    &pb.UUID{Value: req.UserId.Value},
		ProductId: &pb.UUID{Value: req.ProductId.Value},
		Quantity:  req.Quantity,
		Total:     req.Total,
	}

	return &pb.OrderProductResponse{
		Order:   order,
		Message: "Order placed successfully",
	}, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, req *pb.GetUserOrdersRequest) (*pb.GetUserOrdersResponse, error) {
	stmt := `SELECT * from orders WHERE user_id=$1 LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(stmt, req.UserId.Value, req.Limit, req.Page)
	if err != nil {
		return nil, err
	}
	var orders []*pb.Order

	for rows.Next() {
		var order pb.Order
		var userId string
		var productId string
		var orderId string
		err := rows.Scan(
			&orderId,
			&userId,
			&productId,
			&order.Quantity,
			&order.Total,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		order.Id.Value = orderId
		order.ProductId.Value = productId
		order.UserId.Value = userId

		orders = append(orders, &order)
	}

	return &pb.GetUserOrdersResponse{
		Orders:  orders,
		Message: "query successfull",
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	stmt := `SELECT * from orders WHERE id=$1`

	var order pb.Order
	var userId string
	var productId string
	var orderId string
	err := s.db.QueryRow(stmt, req.OrderId.Value).Scan(
		&orderId,
		&userId,
		&productId,
		&order.Quantity,
		&order.Total,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	order.Id.Value = orderId
	order.ProductId.Value = productId
	order.UserId.Value = userId
	return &pb.GetOrderResponse{Order: &order, Message: "query successfull"}, nil
}

func (s *OrderService) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	stmt := `SELECT * from orders LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(stmt, req.Limit, req.Page)
	if err != nil {
		return nil, err
	}
	var orders []*pb.Order

	for rows.Next() {
		var order pb.Order
		var userId string
		var productId string
		var orderId string
		rows.Scan(&orderId, &userId, &productId, &order.Quantity, &order.Total, &order.CreatedAt, &order.UpdatedAt)
		order.Id.Value = orderId
		order.ProductId.Value = productId
		order.UserId.Value = userId

		orders = append(orders, &order)
	}

	return &pb.GetOrdersResponse{
		Orders:  orders,
		Message: "query successfull",
	}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	stmt := `UPDATE orders SET status=$1, quantity=$2, total=$3 WHERE id=$5 RETURNING *`
	var order pb.Order
	var userId string
	var productId string
	var orderId string

	err := s.db.QueryRow(stmt, req.Status, req.Quantity, req.Total, req.OrderId.Value).Scan(
		&orderId,
		&userId,
		&productId,
		&order.Quantity,
		&order.Total,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateOrderResponse{
		Order:   &order,
		Message: "order updated successfully",
	}, nil
}

func (s *OrderService) DelteOrder(ctx context.Context, req *pb.DelteOrderRequest) (*pb.DelteOrderResponse, error) {
	stmt := `DELETE FROM orders WHERE id=$1`
	_, err := s.db.Exec(stmt, req.OrderId.Value)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
