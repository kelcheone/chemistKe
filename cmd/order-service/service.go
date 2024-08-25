package orderservice

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kelcheone/chemistke/internal/database"
	"github.com/kelcheone/chemistke/pkg/codes"
	pb "github.com/kelcheone/chemistke/pkg/grpc/order"
	"github.com/kelcheone/chemistke/pkg/status"
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
		return nil, status.Errorf(codes.Internal, "failed to insert order: %v", err)
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
	stmt := `SELECT id, user_id, product_id, status, quantity, total, created_at, updated_at FROM orders WHERE user_id=$1 LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(stmt, req.UserId.Value, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query orders: %v", err)
	}
	defer rows.Close()

	var orders []*pb.Order

	for rows.Next() {
		var order pb.Order
		var userId, productId, orderId string
		err := rows.Scan(
			&orderId,
			&userId,
			&productId,
			&order.Status,
			&order.Quantity,
			&order.Total,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan order row: %v", err)
		}
		order.Id = &pb.UUID{Value: orderId}
		order.ProductId = &pb.UUID{Value: productId}
		order.UserId = &pb.UUID{Value: userId}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error iterating over rows: %v", err)
	}

	return &pb.GetUserOrdersResponse{
		Orders:  orders,
		Message: "query successful",
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	stmt := `SELECT id, user_id, product_id, status, quantity, total, created_at, updated_at FROM orders WHERE id=$1`
	fmt.Println(stmt)

	var order pb.Order
	var id, userID, productID string
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(ctx, stmt, req.OrderId.Value).Scan(
		&id,
		&userID,
		&productID,
		&order.Status,
		&order.Quantity,
		&order.Total,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "order with ID %s not found", req.OrderId.Value)
		}
		return nil, status.Errorf(codes.Internal, "failed to get order: %v", err)
	}

	order.Id = &pb.UUID{Value: id}
	order.UserId = &pb.UUID{Value: userID}
	order.ProductId = &pb.UUID{Value: productID}
	order.CreatedAt = createdAt.String()
	order.UpdatedAt = updatedAt.String()

	return &pb.GetOrderResponse{Order: &order, Message: "query successful"}, nil
}

func (s *OrderService) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	stmt := `SELECT id, user_id, product_id, status, quantity, total, created_at, updated_at FROM orders LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(stmt, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query orders: %v", err)
	}
	defer rows.Close()

	var orders []*pb.Order

	for rows.Next() {
		var order pb.Order
		var userId, productId, orderId string
		err := rows.Scan(
			&orderId,
			&userId,
			&productId,
			&order.Status,
			&order.Quantity,
			&order.Total,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan order row: %v", err)
		}
		order.Id = &pb.UUID{Value: orderId}
		order.ProductId = &pb.UUID{Value: productId}
		order.UserId = &pb.UUID{Value: userId}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error iterating over rows: %v", err)
	}

	return &pb.GetOrdersResponse{
		Orders:  orders,
		Message: "query successful",
	}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	stmt := `UPDATE orders SET status=$1, quantity=$2, total=$3 WHERE id=$4 RETURNING *`
	var order pb.Order
	var userId, productId, orderId string
	fmt.Println(stmt)
	fmt.Printf("%+v\n", req.Status)

	err := s.db.QueryRow(stmt, req.Status, req.Quantity, req.Total, req.OrderId.Value).Scan(
		&orderId,
		&userId,
		&productId,
		&order.Status,
		&order.Quantity,
		&order.Total,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "order with ID %s not found", req.OrderId.Value)
		}
		return nil, status.Errorf(codes.Internal, "failed to update order: %v", err)
	}

	order.Id = &pb.UUID{Value: orderId}
	order.UserId = &pb.UUID{Value: userId}
	order.ProductId = &pb.UUID{Value: productId}

	return &pb.UpdateOrderResponse{
		Order:   &order,
		Message: "order updated successfully",
	}, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	stmt := `DELETE FROM orders WHERE id=$1`
	result, err := s.db.Exec(stmt, req.OrderId.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete order: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "order with ID %s not found", req.OrderId.Value)
	}

	return &pb.DeleteOrderResponse{
		Message: "order deleted successfully",
	}, nil
}
