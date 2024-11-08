package productservice

import (
	"bytes"
	"context"
	"database/sql"

	"github.com/kelcheone/chemistke/internal/database"
	"github.com/kelcheone/chemistke/internal/files"
	"github.com/kelcheone/chemistke/pkg/codes"
	pb "github.com/kelcheone/chemistke/pkg/grpc/product"
	"github.com/kelcheone/chemistke/pkg/status"
)

type ProductService struct {
	db database.DB
	pb.UnimplementedProductServiceServer
}

func NewProductService(db database.DB) *ProductService {
	return &ProductService{
		db: db,
	}
}

func (s *ProductService) CreateProduct(
	ctx context.Context,
	req *pb.CreateProductRequest,
) (*pb.CreateProductResponse, error) {
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
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(
			codes.Internal,
			"error creating product: %v",
			err,
		)
	}

	return &pb.CreateProductResponse{
		Message: "created user successfully",
		Id: &pb.UUID{
			Value: productId,
		},
	}, nil
}

func (s *ProductService) GetProduct(
	ctx context.Context,
	req *pb.GetProductRequest,
) (*pb.GetProductResponse, error) {
	stmt := `
  SELECT 
   p.id, 
   p.name, 
   p.description, 
   p.category, 
   p.sub_category, 
   p.brand, 
   p.price, 
   p.quantity, 
   pi.url AS image_url,
   pi.image_type
  FROM 
    products p 
  LEFT JOIN 
    productimages pi 
  ON
    p.id  = pi.product_id
  WHERE 
    p.id=$1;`

	rows, err := s.db.Query(stmt, req.Id.Value)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting product: %+v",
			err,
		)
	}

	defer rows.Close()

	var product pb.Product
	var productId string
	product.Images = []*pb.Image{}

	for rows.Next() {

		var imageUrl, imageType sql.NullString

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
			&imageUrl,
			&imageType,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning rows: %v",
				err.Error(),
			)
		}
		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}
		if image.Url != "" {
			product.Images = append(product.Images, &image)
		}

		product.Id = &pb.UUID{Value: productId}

	}
	if productId == "" {
		return nil, status.Errorf(
			codes.NotFound,
			"product with id %s not found.",
			req.Id.Value,
		)
	}
	return &pb.GetProductResponse{
		Product: &product,
	}, nil
}

func (s *ProductService) DeleteProduct(
	ctx context.Context,
	req *pb.DeleteProductRequest,
) (*pb.DeleteProductResponse, error) {
	stmt := `DELETE FROM products WHERE id=$1`
	_, err := s.db.Exec(stmt, req.Id.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(
				codes.NotFound,
				"product with id %s does not exist",
				req.Id.Value,
			)
		}
		return nil, status.Errorf(
			codes.Internal,
			"error deleting product: %v",
			err,
		)
	}
	return &pb.DeleteProductResponse{
		Message: "product deleted successfully",
	}, nil
}

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	req *pb.UpdateProductRequest,
) (*pb.UpdateProductResponse, error) {
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
		if err == sql.ErrNoRows {
			return nil, status.Errorf(
				codes.NotFound,
				"product with id %s not found",
				product.Id.Value,
			)
		}
		return nil, status.Errorf(
			codes.Internal,
			"error updating product: %v",
			err,
		)
	}
	return &pb.UpdateProductResponse{
		Message: "product update successfully",
	}, nil
}

func (s *ProductService) GetProducts(
	ctx context.Context,
	req *pb.GetProductsRequest,
) (*pb.GetProductsResponse, error) {
	stmt := `
	  SELECT
	   p.id,
	   p.name,
	   p.description,
	   p.category,
	   p.sub_category,
	   p.brand,
	   p.price,
	   p.quantity,
	   pi.url AS image_url,
	   pi.image_type
	  FROM
	    products p
	  LEFT JOIN
	    productimages pi
	  ON
	    p.id  = pi.product_id
	  LIMIT $1
	  OFFSET $2
	`

	rows, err := s.db.Query(stmt, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	productsMap := make(map[string]*pb.Product)

	for rows.Next() {
		var product pb.Product
		var productId string
		var imageUrl, imageType sql.NullString

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
			&imageUrl,
			&imageType,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning products: %v",
				err,
			)
		}
		product.Id = &pb.UUID{Value: productId}

		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}

		if exists, found := productsMap[productId]; found {
			if image.Url != "" {
				exists.Images = append(exists.Images, &image)
			}
		} else {
			product.Id = &pb.UUID{Value: productId}
			if image.Url != "" {
				product.Images = []*pb.Image{&image}
			}
			productsMap[productId] = &product
		}

	}

	var products []*pb.Product
	for _, p := range productsMap {
		products = append(products, p)
	}

	return &pb.GetProductsResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

func (s *ProductService) GetProductsByCategory(
	ctx context.Context,
	req *pb.GetProductsByCategoryRequest,
) (*pb.GetProductsByCategoryResponse, error) {
	stmt := `
	  SELECT
	   p.id,
	   p.name,
	   p.description,
	   p.category,
	   p.sub_category,
	   p.brand,
	   p.price,
	   p.quantity,
	   pi.url AS image_url,
	   pi.image_type
	  FROM
	    products p
	  LEFT JOIN
	    productimages pi
	  ON
	    p.id  = pi.product_id
    WHERE 
      category=$1
	  LIMIT $2
	  OFFSET $3
	`
	rows, err := s.db.Query(stmt, req.Category, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	productsMap := make(map[string]*pb.Product)

	for rows.Next() {
		var product pb.Product
		var productId string
		var imageUrl, imageType sql.NullString

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
			&imageUrl,
			&imageType,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning products: %v",
				err,
			)
		}
		product.Id = &pb.UUID{Value: productId}

		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}

		if exists, found := productsMap[productId]; found {
			if image.Url != "" {
				exists.Images = append(exists.Images, &image)
			}
		} else {
			product.Id = &pb.UUID{Value: productId}
			if image.Url != "" {
				product.Images = []*pb.Image{&image}
			}
			productsMap[productId] = &product
		}

	}

	var products []*pb.Product
	for _, p := range productsMap {
		products = append(products, p)
	}

	return &pb.GetProductsByCategoryResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

func (s *ProductService) GetProductsBySubCategory(
	ctx context.Context,
	req *pb.GetProductsBySubCategoryRequest,
) (*pb.GetProductsBySubCategoryResponse, error) {
	stmt := `
	  SELECT
	   p.id,
	   p.name,
	   p.description,
	   p.category,
	   p.sub_category,
	   p.brand,
	   p.price,
	   p.quantity,
	   pi.url AS image_url,
	   pi.image_type
	  FROM
	    products p
	  LEFT JOIN
	    productimages pi
	  ON
	    p.id  = pi.product_id
    WHERE 
     sub_category=$1
	  LIMIT $2
	  OFFSET $3
	`
	rows, err := s.db.Query(stmt, req.SubCategory, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	productsMap := make(map[string]*pb.Product)

	for rows.Next() {
		var product pb.Product
		var productId string
		var imageUrl, imageType sql.NullString

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
			&imageUrl,
			&imageType,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning products: %v",
				err,
			)
		}
		product.Id = &pb.UUID{Value: productId}

		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}

		if exists, found := productsMap[productId]; found {
			if image.Url != "" {
				exists.Images = append(exists.Images, &image)
			}
		} else {
			product.Id = &pb.UUID{Value: productId}
			if image.Url != "" {
				product.Images = []*pb.Image{&image}
			}
			productsMap[productId] = &product
		}

	}

	var products []*pb.Product
	for _, p := range productsMap {
		products = append(products, p)
	}

	return &pb.GetProductsBySubCategoryResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

func (s *ProductService) GetProductsByBrand(
	ctx context.Context,
	req *pb.GetProductsByBrandRequest,
) (*pb.GetProductsByBrandResponse, error) {
	stmt := `
	  SELECT
	   p.id,
	   p.name,
	   p.description,
	   p.category,
	   p.sub_category,
	   p.brand,
	   p.price,
	   p.quantity,
	   pi.url AS image_url,
	   pi.image_type
	  FROM
	    products p
	  LEFT JOIN
	    productimages pi
	  ON
	    p.id  = pi.product_id
    WHERE 
     brand=$1
	  LIMIT $2
	  OFFSET $3
	`
	rows, err := s.db.Query(stmt, req.Brand, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	productsMap := make(map[string]*pb.Product)

	for rows.Next() {
		var product pb.Product
		var productId string
		var imageUrl, imageType sql.NullString

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.SubCategory,
			&product.Brand,
			&product.Price,
			&product.Quantity,
			&imageUrl,
			&imageType,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning products: %v",
				err,
			)
		}
		product.Id = &pb.UUID{Value: productId}

		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}

		if exists, found := productsMap[productId]; found {
			if image.Url != "" {
				exists.Images = append(exists.Images, &image)
			}
		} else {
			product.Id = &pb.UUID{Value: productId}
			if image.Url != "" {
				product.Images = []*pb.Image{&image}
			}
			productsMap[productId] = &product
		}

	}

	var products []*pb.Product
	for _, p := range productsMap {
		products = append(products, p)
	}

	return &pb.GetProductsByBrandResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
	}, nil
}

// upload product image.
func (s *ProductService) UploadProdctImages(
	ctx context.Context,
	req *pb.UploadProdctImagesRequest,
) (*pb.UploadProdctImagesResponse, error) {
	// get product Id, image  then upload it and push the url to the databse.
	fileReader := bytes.NewReader(req.ImageData)

	resp, err := files.UploadImage(
		fileReader,
		req.ProductId.Value,
		req.ImageName,
	)
	// add to database

	stmt := `INSERT INTO productimages(product_id, image_type, url) VALUES ($1, $2, $3) RETURNING id;`
	_, err = s.db.Exec(stmt, req.ProductId.Value, req.ImageType, resp)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"could not upload the image %s",
			err.Error(),
		)
	}

	return &pb.UploadProdctImagesResponse{
		Message: resp,
	}, nil
}

// So to access the images of a given product. enpoint/products/productId/[images...]

func (s *ProductService) GetProductImages(
	ctx context.Context,
	req *pb.GetProductImagesRequest,
) (*pb.GetProductImagesResponse, error) {
	urls, err := s.getProductImages(req.ProductId.Value)
	if err != nil {
		return nil, err
	}

	return &pb.GetProductImagesResponse{
		Urls: urls,
	}, nil
}

func (s *ProductService) getProductImages(
	productId string,
) ([]*pb.Image, error) {
	stmt := `SELECT url, image_type from productimages WHERE product_id=$1;`
	rows, err := s.db.Query(stmt, productId)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting product images: %v",
			err.Error(),
		)
	}

	var urls []*pb.Image

	for rows.Next() {
		var image pb.Image
		err := rows.Scan(&image.Url, &image.ImageType)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning image %v",
				err.Error(),
			)
		}

		urls = append(urls, &image)
	}

	return urls, nil
}
