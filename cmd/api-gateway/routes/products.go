package routes

import (
	"fmt"
	"net/http"

	"github.com/kelcheone/chemistke/cmd/utils"
	product_proto "github.com/kelcheone/chemistke/pkg/grpc/product"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Product struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	SubCategory string  `json:"sub-category"`
	Brand       string  `json:"brand"`
	Price       float32 `json:"price"`
	Quantity    int32   `json:"quantity"`
}

type ProductServer struct {
	ProductClient product_proto.ProductServiceClient
}

func ConnectProductServer(link string) (*ProductServer, func(), error) {
	productConn, err := grpc.NewClient(
		link,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"Faild to connect to the product service: %v",
			err,
		)
	}

	return &ProductServer{
			ProductClient: product_proto.NewProductServiceClient(productConn),
		}, func() {
			productConn.Close()
		}, nil
}

func (p *ProductServer) CreateProduct(c echo.Context) error {
	var product Product

	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	// check claims for the role
	claims := utils.ExtractClaimsFromRequest(c)

	if !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "can't perform this operation.",
		})
	}

	nProduct := &product_proto.CreateProductRequest{
		Product: &product_proto.Product{
			Name:        product.Name,
			Description: product.Description,
			Category:    product.Category,
			SubCategory: product.SubCategory,
			Brand:       product.Brand,
			Price:       product.Price,
			Quantity:    product.Quantity,
		},
	}

	resp, err := p.ProductClient.CreateProduct(c.Request().Context(), nProduct)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (p *ProductServer) GetProduct(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "Invalid request",
		})
	}

	req := &product_proto.GetProductRequest{
		Id: &product_proto.UUID{Value: id},
	}

	resp, err := p.ProductClient.GetProduct(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (p *ProductServer) GetProducts(c echo.Context) error {
	type ProductReq struct {
		Page  int32 `json:"page"`
		Limit int32 `json:"limit"`
	}

	var productReq ProductReq

	if err := c.Bind(&productReq); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	req := &product_proto.GetProductsRequest{
		Limit: productReq.Limit,
		Page:  productReq.Page,
	}

	resp, err := p.ProductClient.GetProducts(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (p *ProductServer) GetProductsByCategory(c echo.Context) error {
	type ProductReq struct {
		Category string `json:"category"`
		Page     int32  `json:"page"`
		Limit    int32  `json:"limit"`
	}

	var productReq ProductReq

	if err := c.Bind(&productReq); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	req := &product_proto.GetProductsByCategoryRequest{
		Category: productReq.Category,
		Limit:    productReq.Limit,
		Page:     productReq.Page,
	}

	resp, err := p.ProductClient.GetProductsByCategory(
		c.Request().Context(),
		req,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (p *ProductServer) GetProductsBySCategory(c echo.Context) error {
	type ProductReq struct {
		SubCategory string `json:"sub-category"`
		Page        int32  `json:"page"`
		Limit       int32  `json:"limit"`
	}

	var productReq ProductReq

	if err := c.Bind(&productReq); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	req := &product_proto.GetProductsBySubCategoryRequest{
		SubCategory: productReq.SubCategory,
		Limit:       productReq.Limit,
		Page:        productReq.Page,
	}

	resp, err := p.ProductClient.GetProductsBySubCategory(
		c.Request().Context(),
		req,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (p *ProductServer) GetProductsByBrand(c echo.Context) error {
	type ProductReq struct {
		Brand string `json:"brand"`
		Page  int32  `json:"page"`
		Limit int32  `json:"limit"`
	}

	var productReq ProductReq

	if err := c.Bind(&productReq); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	req := &product_proto.GetProductsByBrandRequest{
		Brand: productReq.Brand,
		Limit: productReq.Limit,
		Page:  productReq.Page,
	}

	resp, err := p.ProductClient.GetProductsByBrand(
		c.Request().Context(),
		req,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (p *ProductServer) UpdateProduct(c echo.Context) error {
	var product Product

	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	claims := utils.ExtractClaimsFromRequest(c)

	if !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "can't perform this operation.",
		})
	}

	req := &product_proto.UpdateProductRequest{
		Product: &product_proto.Product{
			Id:          &product_proto.UUID{Value: product.Id},
			Name:        product.Name,
			Description: product.Description,
			Category:    product.Category,
			SubCategory: product.SubCategory,
			Brand:       product.Brand,
			Price:       product.Price,
			Quantity:    product.Quantity,
		},
	}

	resp, err := p.ProductClient.UpdateProduct(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, resp)
}

func (p *ProductServer) DeleteProduct(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	claims := utils.ExtractClaimsFromRequest(c)

	if !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "can't perform this operation.",
		})
	}

	fmt.Println("The id is: ", id)
	req := &product_proto.DeleteProductRequest{
		Id: &product_proto.UUID{
			Value: id,
		},
	}

	resp, err := p.ProductClient.DeleteProduct(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, resp)
}
