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
	Id              string                 `json:"id"           example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	Name            string                 `json:"name"         example:"Amoxilin"                             binding:"required"`
	Description     string                 `json:"description"  example:"product description"                  binding:"required"`
	CategoryId      string                 `json:"category_id"  example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"                          binding:"required"`
	SubCategoryId   string                 `json:"sub_category_id" example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"                                 binding:"required"`
	BrandId         string                 `json:"brand_id"        example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"                                  binding:"required"`
	Price           float32                `json:"price"        example:"300.0"                                binding:"required"`
	Quantity        int32                  `json:"quantity"     example:"1000"                                 binding:"required"`
	Featured        bool                   `json:"featured"     example:"false"`
	Images          []*product_proto.Image `json:"images"`
	AverageRating   float32                `json:"average_rating" example:"4.5"`
	ReviewCount     int32                  `json:"review_count" example:"10"`
	CategoryName    string                 `json:"category_name" example:"Antibiotics"`
	SubCategoryName string                 `json:"sub_category_name" example:"Mild"`
	BrandName       string                 `json:"brand_name" example:"J&J"`
}

// Review represents the data required to create a review
type Review struct {
	Id        string  `json:"id"           example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	ProductId string  `json:"product_id"   example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	UserId    string  `json:"user_id"      example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	Rating    float32 `json:"rating"       example:"4.5"`
	Content   string  `json:"content"      example:"Great product"`
	UserName  string  `json:"user_name"    example:"John Doe"`
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
			"faild to connect to the product service: %v",
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
			Name:          product.Name,
			Description:   product.Description,
			CategoryId:    &product_proto.UUID{Value: product.CategoryId},
			SubCategoryId: &product_proto.UUID{Value: product.SubCategoryId},
			BrandId:       &product_proto.UUID{Value: product.BrandId},
			Featured:      product.Featured,
			Price:         product.Price,
			Quantity:      product.Quantity,
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

// a function that takes resp.Product and returns a Product struct
func convertProduct(resp *product_proto.Product) Product {
	return Product{
		Id:              resp.Id.Value,
		Name:            resp.Name,
		Description:     resp.Description,
		Price:           resp.Price,
		CategoryId:      resp.CategoryId.Value,
		SubCategoryId:   resp.SubCategoryId.Value,
		BrandId:         resp.BrandId.Value,
		Images:          resp.Images,
		Featured:        resp.Featured,
		ReviewCount:     resp.ReviewCount,
		CategoryName:    resp.CategoryName,
		SubCategoryName: resp.SubCategoryName,
		BrandName:       resp.BrandName,
		Quantity:        resp.Quantity,
		AverageRating:   resp.AverageRating,
	}
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

	product := convertProduct(resp.Product)

	return c.JSON(http.StatusOK, product)
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

	var products []Product
	for _, product := range resp.Products {
		products = append(products, convertProduct(product))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"products": products,
	})
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
// @Param id path string true "ProductReq Category"
// @Param page query int true "ProductReq Page"
// @Param limit query int true "ProductReq Limit"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/by-category/{id} [get]
func (p *ProductServer) GetProductsByCategory(c echo.Context) error {
	id := c.Param("id")
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

	req := &product_proto.GetProductsByCategoryRequest{
		CategoryId: &product_proto.UUID{Value: id},
		Limit:      int32(n_limit),
		Page:       int32(n_page),
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
	var products []Product
	for _, product := range resp.Products {
		products = append(products, convertProduct(product))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"products": products,
	})
}

// GetProductsBySCategory godoc
// @Summary Get products based on a sub-category
// @Description Get products based on a given sub-category, page, and limit
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "ProductReq SubCategory"
// @Param page query int true "ProductReq Page"
// @Param limit query int true "ProductReq Limit"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/by-subcategory/{id} [get]
func (p *ProductServer) GetProductsBySCategory(c echo.Context) error {
	id := c.Param("id")
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

	req := &product_proto.GetProductsBySubCategoryRequest{
		SubCategoryId: &product_proto.UUID{Value: id},
		Limit:         int32(n_limit),
		Page:          int32(n_page),
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

	var products []Product
	for _, product := range resp.Products {
		products = append(products, convertProduct(product))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"products": products,
	})
}

// GetProducts By Brand godoc
// @Summary Get products based on brand
// @Description Get products based on a given brand name
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "ProductReq Brand"
// @Param page query int true "ProductReq Page"
// @Param limit query int tru "ProductReq Limit"
// @Success 201 {object} Product "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /products/by-brand [get]
func (p *ProductServer) GetProductsByBrand(c echo.Context) error {
	id := c.Param("id")
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

	req := &product_proto.GetProductsByBrandRequest{
		BrandId: &product_proto.UUID{Value: id},
		Limit:   int32(n_limit),
		Page:    int32(n_page),
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

	var products []Product
	for _, product := range resp.Products {
		products = append(products, convertProduct(product))
	}
	return c.JSON(http.StatusOK, map[string]any{
		"products": products,
	})
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
			Id:            &product_proto.UUID{Value: product.Id},
			Name:          product.Name,
			Description:   product.Description,
			CategoryId:    &product_proto.UUID{Value: product.CategoryId},
			SubCategoryId: &product_proto.UUID{Value: product.SubCategoryId},
			BrandId:       &product_proto.UUID{Value: product.BrandId},
			Price:         product.Price,
			Quantity:      product.Quantity,
			Featured:      product.Featured,
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

// CreateReviewRequest
type CreateReviewRequest struct {
	ProductId string  `json:"product_id" example:"1234567890" binding:"required" validate:"required,uuid"`
	UserID    string  `json:"user_id" example:"1234567890" binding:"required" validate:"required,uuid"`
	Title     string  `json:"title" example:"Great product!" binding:"required" validate:"required"`
	Content   string  `json:"content" example:"This product is amazing!" binding:"required" validate:"required"`
	Rating    float32 `json:"rating" example:"4.5" binding:"required" validate:"required"`
}

// Reviews

// CreateReview godoc
// @Summary Create a new review for a product
// @Description Create a new review for a product
// @Tags Products
// @Accept json
// @Produce json
// @Param review body CreateReviewRequest true "Review to create"
// @Success 200 {object} Review
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/reviews [post]
func (p *ProductServer) CreateReview(c echo.Context) error {
	req := &CreateReviewRequest{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: fmt.Sprintf("could not bind request: %+v", err.Error()),
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: fmt.Sprintf("could not validate request: %+v", err.Error()),
		})
	}

	reqProto := &product_proto.CreateReviewRequest{
		ProductId: &product_proto.UUID{Value: req.ProductId},
		UserId:    &product_proto.UUID{Value: req.UserID},
		Title:     req.Title,
		Content:   req.Content,
		Rating:    req.Rating,
	}

	resp, err := p.ProductClient.CreateReview(c.Request().Context(), reqProto)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: err.Error()},
		)
	}

	return c.JSON(http.StatusOK, resp)
}

type ReviewResponse struct {
	Id        string `json:"id"`
	ProductId string `json:"product_id"`
	UserId    string `json:"user_id"`
	Rating    int32  `json:"rating"`
	Title     string `json:"title"`
	UserName  string `json:"user_name"`
	Content   string `json:"content"`
}

// GetReview godoc
// @Summary Get a review for a product
// @Description Get a review for a product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ReviewResponse
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/review/{id} [get]
func (p *ProductServer) GetReview(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "missing product ID",
		})
	}

	reqProto := &product_proto.GetReviewRequest{
		ReviewId: &product_proto.UUID{Value: id},
	}

	resp, err := p.ProductClient.GetReview(c.Request().Context(), reqProto)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: err.Error()},
		)
	}
	// create a new ReviewResponse from the response
	respProto := &ReviewResponse{
		Id:        resp.Review.Id.Value,
		ProductId: resp.Review.ProductId.Value,
		UserId:    resp.Review.UserId.Value,
		Rating:    int32(resp.Review.Rating),
		Title:     resp.Review.Title,
		Content:   resp.Review.Content,
		UserName:  resp.Review.UserName,
	}

	return c.JSON(http.StatusOK, respProto)
}

// GetReviews godoc
// @Summary Get reviews for a product
// @Description Get reviews for a product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ReviewResponse
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/{id}/reviews [get]
func (p *ProductServer) GetReviews(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "missing product ID",
		})
	}

	reqProto := &product_proto.GetReviewsRequest{
		ProductId: &product_proto.UUID{Value: id},
	}

	resp, err := p.ProductClient.GetReviews(c.Request().Context(), reqProto)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: err.Error()},
		)
	}
	var reviews []*ReviewResponse
	for _, review := range resp.Reviews {
		reviews = append(reviews, &ReviewResponse{
			Id:        review.Id.Value,
			ProductId: review.ProductId.Value,
			UserId:    review.UserId.Value,
			Rating:    int32(review.Rating),
			Title:     review.Title,
			Content:   review.Content,
			UserName:  review.UserName,
		})
	}

	num_reviews := len(reviews)
	sum_ratings := int32(0)
	avg_rating := float32(0)
	if num_reviews > 0 {
		sum_ratings = int32(num_reviews)
		avg_rating = float32(sum_ratings) / float32(num_reviews)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"reviews":        reviews,
		"total_ratings":  num_reviews,
		"average_rating": avg_rating,
	})
}

// GetProductRating godoc
// @Summary Get rating for a product
// @Description Get rating for a product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ReviewResponse
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/rating/{id} [get]
func (p *ProductServer) GetProductRating(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "missing product ID",
		})
	}

	reqProto := &product_proto.GetProductRatingRequest{
		ProductId: &product_proto.UUID{Value: id},
	}

	log.Printf("The productId is: %s ", reqProto.ProductId.Value)

	resp, err := p.ProductClient.GetProductRating(c.Request().Context(), reqProto)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: err.Error()},
		)
	}
	return c.JSON(http.StatusOK, resp)
}

// ProductCategory represents a category of products
type ProductCategory struct {
	Id          string `json:"id"           example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	Name        string `json:"name"         example:"Amoxilin"                             binding:"required"`
	Description string `json:"description"  example:"product description"                  binding:"required"`
	Featured    bool   `json:"featured"     example:"true" binding:"required"`
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new category
// @Tags Products
// @Accept json
// @Produce json
// @Param category body ProductCategory true "Category"
// @Success 200 {object} ProductCategory
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/categories [post]
func (p *ProductServer) CreateCategory(c echo.Context) error {
	category := &ProductCategory{}
	if err := c.Bind(category); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}
	createCategoryReq := &product_proto.CreateCategoryRequest{
		Name:        category.Name,
		Description: category.Description,
		Featured:    category.Featured,
	}

	resp, err := p.ProductClient.CreateCategory(c.Request().Context(), createCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusCreated, resp)
}

// GetCategory godoc
// @Summary Get A category by id
// @Description Get a category by id
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} ProductCategory
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/categories/{id} [get]
func (p *ProductServer) GetCategory(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id is required",
		})
	}
	getCategoryReq := &product_proto.GetCategoryRequest{
		Id: &product_proto.UUID{Value: id},
	}

	resp, err := p.ProductClient.GetCategory(c.Request().Context(), getCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	categoryResp := &ProductCategory{
		Id:          resp.Category.Id.Value,
		Name:        resp.Category.Name,
		Description: resp.Category.Description,
		Featured:    resp.Category.Featured,
	}

	return c.JSON(http.StatusOK, categoryResp)
}

// UpdateCategory godoc
// @Summary Update a category by id
// @Description Update a category by id
// @Tags Products
// @Accept json
// @Produce json
// @Param category body ProductCategory true "Category"
// @Success 200 {object} ProductCategory
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/categories [patch]
func (p *ProductServer) UpdateCategory(c echo.Context) error {
	category := &ProductCategory{}
	if err := c.Bind(category); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}
	updateCategoryReq := &product_proto.UpdateCategoryRequest{
		Id:          &product_proto.UUID{Value: category.Id},
		Name:        category.Name,
		Description: category.Description,
		Featured:    category.Featured,
	}

	resp, err := p.ProductClient.UpdateCategory(c.Request().Context(), updateCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

// DeleteCategory godoc
// @Summary Deletes a given category.
// @Description Deletes a given category.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 204
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/categories/{id} [delete]
func (p *ProductServer) DeleteCategory(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id is required",
		})
	}
	deleteCategoryReq := &product_proto.DeleteCategoryRequest{
		Id: &product_proto.UUID{Value: id},
	}

	_, err := p.ProductClient.DeleteCategory(c.Request().Context(), deleteCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetCategories godoc
// @Summary Gets product categories.
// @Description Gets paginated product Categories
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} product_proto.GetCategoriesResponse
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/categories [get]
func (p *ProductServer) GetCategories(c echo.Context) error {
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	intPage, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid page",
		})
	}

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid limit",
		})
	}

	getCategoriesReq := &product_proto.GetCategoriesRequest{
		Limit:  int32(intLimit),
		Offset: int32(intPage),
	}

	resp, err := p.ProductClient.GetCategories(c.Request().Context(), getCategoriesReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	var categories []ProductCategory
	for _, category := range resp.Categories {
		categories = append(categories, ProductCategory{
			Id:          category.Id.Value,
			Name:        category.Name,
			Description: category.Description,
			Featured:    category.Featured,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"categories": categories,
	})
}

// GetFeaturedCategories godoc
// @Summary Gets featured product categories.
// @Description Gets paginated featured product Categories
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} product_proto.GetCategoriesResponse
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/categories/featured [get]
func (p *ProductServer) GetFeaturedCategories(c echo.Context) error {
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	intPage, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid page format",
		})
	}

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid page format",
		})
	}
	req := &product_proto.GetFeaturedCategoriesRequest{
		Limit:  int32(intLimit),
		Offset: int32(intPage),
	}

	resp, err := p.ProductClient.GetFeaturedCategories(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: "failed to get featured categories",
		})
	}

	var categories []ProductCategory
	for _, category := range resp.Categories {
		categories = append(categories, ProductCategory{
			Id:          category.Id.Value,
			Name:        category.Name,
			Description: category.Description,
			Featured:    category.GetFeatured(),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"categories": categories,
	})
}

// SubCategories
type ProductSubCategory struct {
	Id          string `json:"id"           example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	CategoryId  string `json:"category_id"  example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8" binding:"required"`
	Name        string `json:"name"         example:"anticonvulsants" binding:"required"`
	Description string `json:"description"  example:"product description" binding:"required"`
}

// CreateSubCategory godoc
// @Summary Creates a subcategory within a category
// @Description Creates a new subcategory within a specified category.
// @Tags Products
// @Accept json
// @Produce json
// @Param subcategory body ProductSubCategory true "Subcategory"
// @Success 201 {object} ProductSubCategory
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/subcategories [post]
func (p *ProductServer) CreateSubCategory(c echo.Context) error {
	var subcategory ProductSubCategory
	if err := c.Bind(&subcategory); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid subcategory",
		})
	}

	createSubCategoryReq := &product_proto.CreateSubCategoryRequest{
		CategoryId:  &product_proto.UUID{Value: subcategory.CategoryId},
		Name:        subcategory.Name,
		Description: subcategory.Description,
	}

	resp, err := p.ProductClient.CreateSubCategory(c.Request().Context(), createSubCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

// GetSubCategory godoc
// @Summary Retrieves a subcategory by its ID
// @Description Retrieves a subcategory by its ID.
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Subcategory ID"
// @Success 200 {object} ProductSubCategory
// @Failure 400 {object} ErrResponse
// @Failure 404 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/subcategories/{id} [get]
func (p *ProductServer) GetSubCategory(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid subcategory id",
		})
	}

	getSubCategoryReq := &product_proto.GetSubCategoryRequest{
		Id: &product_proto.UUID{Value: id},
	}

	resp, err := p.ProductClient.GetSubCategory(c.Request().Context(), getSubCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, ProductSubCategory{
		Id:          resp.SubCategory.Id.Value,
		CategoryId:  resp.SubCategory.CategoryId.Value,
		Name:        resp.SubCategory.Name,
		Description: resp.SubCategory.Description,
	})
}

// UpdateSubCategory godoc
// @Summary Updates a subcategory by its ID
// @Description Updates a subcategory by its ID.
// @Tags Products
// @Accept json
// @Produce json
// @Param subcategory body ProductSubCategory true "Subcategory"
// @Success 200 {object} ProductSubCategory
// @Failure 400 {object} ErrResponse
// @Failure 404 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/subcategories [patch]
func (p *ProductServer) UpdateSubCategory(c echo.Context) error {
	subcategory := &ProductSubCategory{}
	if err := c.Bind(subcategory); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	updateSubCategoryReq := &product_proto.UpdateSubCategoryRequest{
		Id:          &product_proto.UUID{Value: subcategory.Id},
		Name:        subcategory.Name,
		Description: subcategory.Description,
	}

	resp, err := p.ProductClient.UpdateSubCategory(c.Request().Context(), updateSubCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteSubCategory godoc
// @Summary Deletes a subcategory by its ID
// @Descrption Deletes a subcategory by its ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Subcategory ID"
// @Success 204
// @Failure 400 {object} ErrResponse
// @Failure 404 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/subcategories/{id} [delete]
func (p *ProductServer) DeleteSubCategory(c echo.Context) error {
	id := c.Param("id")

	deleteSubCategoryReq := &product_proto.DeleteSubCategoryRequest{
		Id: &product_proto.UUID{Value: id},
	}

	_, err := p.ProductClient.DeleteSubCategory(c.Request().Context(), deleteSubCategoryReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetSubCategories godoc
// @Summary Retrives paginated Subcategories
// @Description Gets subcategories in a paginated manner
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param id path string true "Category ID"
// @Success 200 {array} ProductSubCategory
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/categories/{id}/subcategories [get]
func (p *ProductServer) GetSubCategories(c echo.Context) error {
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	category := c.Param("id")

	if category == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "Category ID is required",
		})
	}

	intPage, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	getSubCategoriesReq := &product_proto.GetSubCategoriesRequest{
		CategoryId: &product_proto.UUID{Value: category},
		Limit:      int32(intLimit),
		Offset:     int32(intPage),
	}

	resp, err := p.ProductClient.GetSubCategories(c.Request().Context(), getSubCategoriesReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	var subCategories []ProductSubCategory

	for _, subCategory := range resp.SubCategories {
		subCategories = append(subCategories, ProductSubCategory{
			Id:          subCategory.Id.Value,
			Name:        subCategory.Name,
			Description: subCategory.Description,
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"subCategories": subCategories,
	})
}

// Brand

type Brand struct {
	Id          string `json:"id"           example:"f183e73c-687d-44ad-83e6-636ecbb7a7d8"`
	Name        string `json:"name"         example:"J&J" binding:"required"`
	Description string `json:"description"  example:"Johnson and Johnson"`
}

// CreateBrand godoc
// @Summary Creates a new brand
// @Description Creates a new brand
// @Tags Products
// @Accept json
// @Produce json
// @Param brand body Brand true "Brand"
// @Success 200 {object} Brand
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/brands [post]
func (p *ProductServer) CreateBrand(c echo.Context) error {
	var brand Brand
	if err := c.Bind(&brand); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	createBrandReq := &product_proto.CreateBrandRequest{
		Name:        brand.Name,
		Description: brand.Description,
	}

	resp, err := p.ProductClient.CreateBrand(c.Request().Context(), createBrandReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetBrand godoc
// @Summary Gets a brand by id
// @Description Gets a brand by id
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Brand ID"
// @Success 200 {object} Brand
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/brands/{id} [get]
func (p *ProductServer) GetBrand(c echo.Context) error {
	id := c.Param("id")

	getBrandReq := &product_proto.GetBrandRequest{
		Id: &product_proto.UUID{Value: id},
	}

	resp, err := p.ProductClient.GetBrand(c.Request().Context(), getBrandReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	brandResp := &Brand{
		Id:          resp.Brand.Id.Value,
		Name:        resp.Brand.Name,
		Description: resp.Brand.Description,
	}

	return c.JSON(http.StatusOK, brandResp)
}

// UpdateBrand godoc
// @Summary Updates a brand by id
// @Description Updates a brand by id
// @Tags Products
// @Accept json
// @Produce json
// @Param brand body Brand true "Brand"
// @Success 200 {object} Brand
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/brands [patch]
func (p *ProductServer) UpdateBrand(c echo.Context) error {
	var brand Brand
	if err := c.Bind(&brand); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	updateBrandReq := &product_proto.UpdateBrandRequest{
		Id:          &product_proto.UUID{Value: brand.Id},
		Name:        brand.Name,
		Description: brand.Description,
	}

	resp, err := p.ProductClient.UpdateBrand(c.Request().Context(), updateBrandReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// DeleteBrand godoc
// @Summary Deletes a brand by id
// @Description Deletes a brand by id
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Brand ID"
// @Success 200 {object} Brand
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Security BearerAuth
// @Router /products/brands/{id} [delete]
func (p *ProductServer) DeleteBrand(c echo.Context) error {
	id := c.Param("id")

	deleteBrandReq := &product_proto.DeleteBrandRequest{
		Id: &product_proto.UUID{Value: id},
	}

	_, err := p.ProductClient.DeleteBrand(c.Request().Context(), deleteBrandReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// Get Brands godoc
// @Summary Gets all brands
// @Description Gets all brands
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Limit"
// @Success 200 {object} Brand
// @Failure 400 {object} ErrResponse
// @Failure 500 {object} ErrResponse
// @Router /products/brands [get]
func (p *ProductServer) GetBrands(c echo.Context) error {
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	intPage, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	getBrandsReq := &product_proto.GetBrandsRequest{
		Limit:  int32(intLimit),
		Offset: int32(intPage),
	}

	resp, err := p.ProductClient.GetBrands(c.Request().Context(), getBrandsReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}
