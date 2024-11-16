package routes

import (
	"fmt"
	"net/http"

	"github.com/kelcheone/chemistke/cmd/utils"
	order_proto "github.com/kelcheone/chemistke/pkg/grpc/order"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderServer struct {
	OrderClient order_proto.OrderServiceClient
}

type Order struct {
	Id        string  `json:"id"`
	ProductId string  `json:"product_id"`
	UserId    string  `json:"user_id"`
	Status    string  `json:"status"`
	Quantity  int32   `json:"quantity"`
	Total     float32 `json:"total"`
}

type IdReq struct {
	Id string `json:"id"`
}

func ConnectOrdersServer(link string) (*OrderServer, func(), error) {
	orderConn, err := grpc.NewClient(
		link,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not connect to order server: %v",
			err,
		)
	}

	conn := &OrderServer{
		OrderClient: order_proto.NewOrderServiceClient(orderConn),
	}

	return conn, func() {
		orderConn.Close()
	}, nil
}

func (o *OrderServer) CreateOrder(c echo.Context) error {
	var order Order

	if err := c.Bind(&order); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	nOrder := &order_proto.OrderProductRequest{
		ProductId: &order_proto.UUID{Value: order.ProductId},
		UserId:    &order_proto.UUID{Value: order.UserId},
		Quantity:  order.Quantity,
		Total:     float32(order.Total),
	}
	resp, err := o.OrderClient.OrderProduct(c.Request().Context(), nOrder)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (o *OrderServer) GetOrder(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid id",
		})
	}

	resp, err := o.OrderClient.GetOrder(
		c.Request().Context(),
		&order_proto.GetOrderRequest{OrderId: &order_proto.UUID{Value: id}},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

type PaginatedReq struct {
	Id    string `json:"id"`
	Page  int32  `json:"page"`
	Limit int32  `json:"limit"`
}

func (o *OrderServer) GetUserOders(c echo.Context) error {
	var req PaginatedReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid Request",
		})
	}

	claims := utils.ExtractClaimsFromRequest(c)
	if claims.Id != req.Id || !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized to perform this action",
		})
	}

	nReq := &order_proto.GetUserOrdersRequest{
		UserId: &order_proto.UUID{Value: req.Id},
		Limit:  req.Limit,
		Page:   req.Page,
	}

	resp, err := o.OrderClient.GetUserOrders(c.Request().Context(), nReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (o *OrderServer) GetOders(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized to perform this action",
		})
	}

	var req PaginatedReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	nReq := &order_proto.GetOrdersRequest{
		Limit: req.Limit,
		Page:  req.Page,
	}

	resp, err := o.OrderClient.GetOrders(c.Request().Context(), nReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (o *OrderServer) UpdateOrder(c echo.Context) error {
	var order Order

	if err := c.Bind(&order); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	claims := utils.ExtractClaimsFromRequest(c)
	if claims.Id != order.UserId || !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized to perform this action",
		})
	}

	nORder := &order_proto.UpdateOrderRequest{
		OrderId:  &order_proto.UUID{Value: order.Id},
		Status:   order.Status,
		Quantity: order.Quantity,
		Total:    order.Total,
	}

	resp, err := o.OrderClient.UpdateOrder(c.Request().Context(), nORder)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, resp)
}

func (o *OrderServer) DeleteOrder(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid delete request",
		})
	}
	order, err := o.OrderClient.GetOrder(
		c.Request().Context(),
		&order_proto.GetOrderRequest{OrderId: &order_proto.UUID{Value: id}},
	)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrResponse{
			Message: err.Error(),
		})
	}

	claims := utils.ExtractClaimsFromRequest(c)
	if claims.Id != order.Order.UserId.Value || !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized to perform this action",
		})
	}

	resp, err := o.OrderClient.DeleteOrder(
		c.Request().Context(),
		&order_proto.DeleteOrderRequest{
			OrderId: &order_proto.UUID{Value: id},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusNoContent, resp)
}
