package productservice

import (
	"context"
	"fmt"
	"time"

	"github.com/kelcheone/chemistke/internal/database"
	"github.com/kelcheone/chemistke/internal/files"
	pb "github.com/kelcheone/chemistke/pkg/grpc/product"
)

type ProductService struct {
	db    database.DB
	files *files.Files
	pb.UnimplementedProductServiceServer
}

func NewProductService(files *files.Files, db database.DB) *ProductService {
	return &ProductService{
		db:    db,
		files: files,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	stmt := `INSERT INTO products (name,description, category, sub_category, brand, price, quantity) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	product := req.Product
	var productId string
	err := s.db.QueryRow(stmt,
		product.Name,
		product.Description,
		product.Category,
		product.SubCategory,
		product.Brand,
		product.Price,
		product.Quantity,
	).Scan(&productId)
	if err != nil {
		return nil, err
	}

	return &pb.CreateProductResponse{
		Message: "created user successfully",
		Id: &pb.UUID{
			Value: productId,
		},
	}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	stmt := `SELECT id, name, description, category, sub_category, brand, price, quantity FROM products WHERE id=$1`
	var product pb.Product
	var productId string
	err := s.db.QueryRow(stmt, req.Id.Value).Scan(
		&productId,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.SubCategory,
		&product.Brand,
		&product.Price,
		&product.Quantity,
	)
	if err != nil {
		return nil, err
	}
	product.Id = &pb.UUID{Value: productId}
	return &pb.GetProductResponse{
		Product: &product,
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	stmt := `DELETE FROM products WHERE id=$1`
	_, err := s.db.Exec(stmt, req.Id.Value)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteProductResponse{Message: "product deleted successfully"}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	stmt := `UPDATE products SET name=$1, description=$2, category=$3, sub_category=$4, brand=$5, price=$6, quantity=$7 WHERE id=$8`
	product := req.Product

	_, err := s.db.Exec(stmt,
		product.Name,
		product.Description,
		product.Category,
		product.SubCategory,
		product.Brand,
		product.Price,
		product.Quantity,
		product.Id.Value,
	)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateProductResponse{
		Message: "product update successfully",
	}, nil
}

func (s *ProductService) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	stmt := `SELECT id, name, description, category, sub_category, brand, price, quantity FROM products LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(stmt, req.Limit, req.Page)
	if err != nil {
		return nil, err
	}

	var products []*pb.Product
	for rows.Next() {
		var product pb.Product
		var productId string

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
		)
		if err != nil {
			return nil, err
		}
		product.Id = &pb.UUID{Value: productId}

		products = append(products, &product)
	}

	return &pb.GetProductsResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, req *pb.GetProductsByCategoryRequest) (*pb.GetProductsByCategoryResponse, error) {
	stmt := `SELECT id, name, description, category, sub_category, brand, price, quantity FROM products WHERE category=$1 LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(stmt, req.Category, req.Limit, req.Page)
	if err != nil {
		return nil, err
	}

	var products []*pb.Product
	for rows.Next() {
		var product pb.Product
		var productId string

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
		)
		if err != nil {
			return nil, err
		}
		product.Id = &pb.UUID{Value: productId}

		products = append(products, &product)
	}

	return &pb.GetProductsByCategoryResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

func (s *ProductService) GetProductsBySubCategory(ctx context.Context, req *pb.GetProductsBySubCategoryRequest) (*pb.GetProductsBySubCategoryResponse, error) {
	stmt := `SELECT id, name, description, category, sub_category, brand, price, quantity FROM products WHERE sub_category=$1 LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(stmt, req.SubCategory, req.Limit, req.Page)
	if err != nil {
		return nil, err
	}

	var products []*pb.Product
	for rows.Next() {
		var product pb.Product
		var productId string

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
		)
		if err != nil {
			return nil, err
		}
		product.Id = &pb.UUID{Value: productId}

		products = append(products, &product)
	}

	return &pb.GetProductsBySubCategoryResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

func (s *ProductService) GetProductsByBrand(ctx context.Context, req *pb.GetProductsByBrandRequest) (*pb.GetProductsByBrandResponse, error) {
	stmt := `SELECT id, name, description, category, sub_category, brand, price, quantity FROM products WHERE brand=$1 LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(stmt, req.Brand, req.Limit, req.Page)
	if err != nil {
		return nil, err
	}

	var products []*pb.Product
	for rows.Next() {
		var product pb.Product
		var productId string

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
		)
		if err != nil {
			return nil, err
		}
		product.Id = &pb.UUID{Value: productId}

		products = append(products, &product)
	}

	return &pb.GetProductsByBrandResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

// GetUploadURL expects the productId, userId and the filename and returns the required file name.
func (s *ProductService) GetUploadURL(ctx context.Context, req *pb.GetUploadURLRequest) (*pb.GetUploadURLResponse, error) {
	randId := time.Now().UnixNano()
	key := fmt.Sprintf("products/%s/%d_%s", req.Id.Value, randId, req.FileName)
	url, err := s.files.GetPresignedURL(key)
	if err != nil {
		return nil, err
	}
	return &pb.GetUploadURLResponse{
		Url: url,
	}, nil
}

// So to access the images of a given product. enpoint/products/productId/[images...]

func (s *ProductService) GetProductImages(ctx context.Context, req *pb.GetProductImagesRequest) (*pb.GetProductImagesResponse, error) {
	// the images are found in the bucketname/products/images
	images, err := s.files.GetProductImages(ctx, req.ProductId.Value)
	if err != nil {
		return nil, err
	}

	return &pb.GetProductImagesResponse{
		Urls: images,
	}, nil

}
