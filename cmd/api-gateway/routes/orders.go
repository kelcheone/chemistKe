package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/kelcheone/chemistke/cmd/utils"
	order_proto "github.com/kelcheone/chemistke/pkg/grpc/order"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderServer struct {
	OrderClient order_proto.OrderServiceClient
}

// Order represents the data required and returned by the Order Service endpoints
type Order struct {
	Id        string  `json:"id"         example:"62e9e179-3aaa-4dd5-a098-21f20da10f90"`
	ProductId string  `json:"product_id" example:"62e9e179-3aaa-4dd5-a098-21f20da10f90" binding:"required"`
	UserId    string  `json:"user_id"    example:"62e9e179-3aaa-4dd5-a098-21f20da10f90" binding:"required"`
	Status    string  `json:"status"     example:"pending"`
	Quantity  int32   `json:"quantity"   example:"10"                                   binding:"required"`
	Total     float32 `json:"total"      example:"100"                                  binding:"required"`
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

// CreateOrder godoc
// @Summary create an order in the system
// @Description create a new order.
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body Order true "Oder info to create"
// @Success 201 {Object} Order "Successfly created a product"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /orders [post]
func (o *OrderServer) CreateOrder(c echo.Context) error {
	var order Order

	if err := c.Bind(&order); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	// TODO: Calculate the total based on the price of a product

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

// GetOrder godoc
// @Summary Get an order by id
// @Description Get order by id
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path string true "Oder ID"
// @Success 200 {Object} Order "Successfly fetched order"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /orders/{id} [get]
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

// GetUserOrder godoc
// @Summary Get a users' orders
// @Description Get orders for a given user
// @Tags Orders
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Param page query int true "PaginatedReq Page"
// @Param limit query int true "PaginatedReq limit"
// @Success 201 {Object} Order "Successfly fetched user orders"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /orders/user [get]
func (o *OrderServer) GetUserOders(c echo.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}
	page := c.QueryParam("page")

	n_page, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	limit := c.QueryParam("limit")

	n_limit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}
	claims := utils.ExtractClaimsFromRequest(c)
	if claims.Id != id || !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized to perform this action",
		})
	}

	log.Println("--------------------------------------------------------")
	log.Println("id: ", id)

	nReq := &order_proto.GetUserOrdersRequest{
		UserId: &order_proto.UUID{Value: id},
		Limit:  int32(n_limit),
		Page:   int32(n_page),
	}

	resp, err := o.OrderClient.GetUserOrders(c.Request().Context(), nReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// GetOrders godoc
// @Summary Get  orders
// @Description Get orders
// @Tags Orders
// @Accept json
// @Produce json
// @Param page query int true "PaginatedReq Page"
// @Param limit query int true "PaginatedReq limit"
// @Success 201 {Object} Order "Successfly fetched orders"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /orders [get]
func (o *OrderServer) GetOders(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized to perform this action",
		})
	}

	page := c.QueryParam("page")

	n_page, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	limit := c.QueryParam("limit")

	n_limit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	nReq := &order_proto.GetOrdersRequest{
		Limit: int32(n_limit),
		Page:  int32(n_page),
	}

	resp, err := o.OrderClient.GetOrders(c.Request().Context(), nReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// UpdateOrder godoc
// @Summary update a given order
// @Description update an order.
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body Order true "Oder info to create"
// @Success 201 {Object} Order  "Oder Successfly updated"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /orders [patch]
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

// DeleteOrder godoc
// @Summary Delete order
// @Description Delete order by id
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path string true "Order Id"
// @Success 201 {Object} Order "Successfly deleted order"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /orders/{id} [delete]
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
