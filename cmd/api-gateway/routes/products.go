package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/kelcheone/chemistke/cmd/utils"
	product_proto "github.com/kelcheone/chemistke/pkg/grpc/product"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Product represents the data required to create a product
type Product struct {
	Id          string  `json:"id"           example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	Name        string  `json:"name"         example:"Amoxilin"                             binding:"required"`
	Description string  `json:"description"  example:"product description"                  binding:"required"`
	Category    string  `json:"category"     example:"Antibiotics"                          binding:"required"`
	SubCategory string  `json:"sub-category" example:"Mild"                                 binding:"required"`
	Brand       string  `json:"brand"        example:"J&J"                                  binding:"required"`
	Price       float32 `json:"price"        example:"300.0"                                binding:"required"`
	Quantity    int32   `json:"quantity"     example:"1000"                                 binding:"required"`
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

// CreateProduct godoc
// @Summary Creates a new Product
// @Description Create a new product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body Product true "Product information to create"
// @Success 201 {object} Product "Successfully updated product"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /products [post]
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

// GetProduct godoc
// @Summary Get product by ID
// @Description Get a product by product ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product Id"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/{id} [get]
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

// GetProducts godoc
// @Summary Get products
// @Description Get products based on page and limit
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int true "PaginationRequest Page"
// @Param limit query int tru "PaginationRequest Limit"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products [get]
func (p *ProductServer) GetProducts(c echo.Context) error {
	var productReq PaginationRequest

	page := c.QueryParam("page")

	n_page, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}
	productReq.Page = n_page

	limit := c.QueryParam("limit")

	n_limit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	productReq.Limit = n_limit

	req := &product_proto.GetProductsRequest{
		Limit: int32(productReq.Limit),
		Page:  int32(productReq.Page),
	}

	resp, err := p.ProductClient.GetProducts(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// ProductReq represents the required product request by the endpoint
type ProductReq struct {
	Category    string `json:"category"     example:"Antibiotics"`
	SubCategory string `json:"sub-category"`
	Brand       string `json:"brand"`
	Page        int32  `json:"page"`
	Limit       int32  `json:"limit"`
}

// GetProductsByCategory godoc
// @Summary Get products Based on a category
// @Description Get products based on a givne Category
// @Tags Products
// @Accept json
// @Produce json
// @Param category query string true "ProductReq Category"
// @Param page query int true "ProductReq Page"
// @Param limit query int true "ProductReq Limit"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/get-products-by-category [get]
func (p *ProductServer) GetProductsByCategory(c echo.Context) error {
	category := c.QueryParam("category")
	if category == "" {
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

	req := &product_proto.GetProductsByCategoryRequest{
		Category: category,
		Limit:    int32(n_limit),
		Page:     int32(n_page),
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

// GetProductsBySCategory godoc
// @Summary Get products based on a sub-category
// @Description Get products based on a given sub-category, page, and limit
// @Tags Products
// @Accept json
// @Produce json
// @Param sub-category query string true "ProductReq SubCategory"
// @Param page query int true "ProductReq Page"
// @Param limit query int tru "ProductReq Limit"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/get-products-by-sub-category [get]
func (p *ProductServer) GetProductsBySCategory(c echo.Context) error {
	subCategory := c.QueryParam("sub-category")
	if subCategory == "" {
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

	req := &product_proto.GetProductsBySubCategoryRequest{
		SubCategory: subCategory,
		Limit:       int32(n_limit),
		Page:        int32(n_page),
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

// GetProducts By Brand godoc
// @Summary Get products based on brand
// @Description Get products based on a given brand name
// @Tags Products
// @Accept json
// @Produce json
// @Param brand query string true "ProductReq Brand"
// @Param page query int true "ProductReq Page"
// @Param limit query int tru "ProductReq Limit"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/get-products-by-brand [get]
func (p *ProductServer) GetProductsByBrand(c echo.Context) error {
	brand := c.QueryParam("brand")
	if brand == "" {
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

	req := &product_proto.GetProductsByBrandRequest{
		Brand: brand,
		Limit: int32(n_limit),
		Page:  int32(n_page),
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

// UpdateProduct godoc
// @Summary Update a  product
// @Description update a given product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body Product true "Product information to create"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /products [patch]
func (p *ProductServer) UpdateProduct(c echo.Context) error {
	var product Product

	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	claims := utils.ExtractClaimsFromRequest(c)

	log.Printf("User Role: %v", claims.Admin)

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

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product based on a given id
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product id"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /products/{id} [delete]
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

// UploadImage godoc
// @Summary Upload a product image
// @Description Upload an image for a given product using multipart/form-data.
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Param product-id formData string true "Product ID"
// @Param image-type formData string false "Type of the image (e.g. thumbnail, banner, general)"
// @Param file formData file true "Image file to upload"
// @Success 200 {object} product_proto.UploadProdctImagesResponse "Successfully uploaded image"
// @Failure 400 {object} ErrResponse "Invalid input data"
// @Failure 500 {object} ErrResponse "Internal server error"
// @Security BearerAuth
// @Router /products/images/upload [post]
func (p *ProductServer) UploadImage(c echo.Context) error {
	productId := c.FormValue("product-id")
	imgType := c.FormValue("image-type")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: "invalid request",
		})
	}
	if imgType == "" {
		imgType = "general"
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: "could not open file",
		})
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: "could not read file",
		})
	}

	resp, err := p.ProductClient.UploadProdctImages(
		c.Request().Context(),
		&product_proto.UploadProdctImagesRequest{
			ProductId: &product_proto.UUID{Value: productId},
			ImageData: fileBytes,
			ImageType: imgType,
			ImageName: fileHeader.Filename,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: fmt.Sprintf("could not upload: %+v", err.Error()),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetProductImages godoc
// @Summary  get a product's images
// @Description Get the images of a given product id
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product id"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/images/{id} [get]
func (p *ProductServer) GetProductImages(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid id",
		})
	}

	req := &product_proto.GetProductImagesRequest{
		ProductId: &product_proto.UUID{Value: id},
	}

	resp, err := p.ProductClient.GetProductImages(c.Request().Context(), req)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: err.Error()},
		)
	}

	return c.JSON(http.StatusOK, resp)
}
