package productservice

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/url"
	"time"

	"github.com/kelcheone/chemistke/internal/database"
	"github.com/kelcheone/chemistke/internal/files"
	"github.com/kelcheone/chemistke/pkg/codes"
	pb "github.com/kelcheone/chemistke/pkg/grpc/product"
	"github.com/kelcheone/chemistke/pkg/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	stmt := `INSERT INTO products (name,description, category_id, sub_category_id, brand_id, price, quantity, featured) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	product := req.Product
	var productId string
	err := s.db.QueryRow(stmt,
		product.Name,
		product.Description,
		product.CategoryId.Value,
		product.SubCategoryId.Value,
		product.BrandId.Value,
		product.Price,
		product.Quantity,
		product.Featured,
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
  p.category_id,
  p.sub_category_id,
  p.brand_id,
  p.price,
  p.quantity,
  p.featured,
  p.slug,
  pc.name AS category_name,
  psc.name AS sub_category_name,
  pb.name AS brand_name,
  pi.url AS image_url,
  pi.image_type,
  pr.review_count,
  pr.average_rating,
  p.created_at,
  p.updated_at
FROM
  products p
LEFT JOIN (
  SELECT
    product_id,
    COUNT(*) AS review_count,
    AVG(rating) AS average_rating
  FROM
    product_reviews
  GROUP BY
    product_id
) pr ON p.id = pr.product_id
LEFT JOIN
  productimages pi ON p.id = pi.product_id
LEFT JOIN
 product_category pc ON p.category_id = pc.id
LEFT JOIN
  product_sub_category psc ON p.sub_category_id = psc.id
LEFT JOIN
  product_brand pb ON p.brand_id = pb.id
WHERE
  p.id = $1;`
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
		var reviewCount sql.NullInt32
		var averageRating sql.NullFloat64
		var categoryName, subCategoryName, brandName sql.NullString
		var createdAt, updatedAt sql.NullTime
		var brandId, categoryId, subCategoryId string

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&categoryId,
			&subCategoryId,
			&brandId,
			&product.Price,
			&product.Quantity,
			&product.Featured,
			&product.Slug,
			&categoryName,
			&subCategoryName,
			&brandName,
			&imageUrl,
			&imageType,
			&reviewCount,
			&averageRating,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning rows: %v",
				err.Error(),
			)
		}
		product.CategoryId = &pb.UUID{Value: categoryId}
		product.SubCategoryId = &pb.UUID{Value: subCategoryId}
		product.BrandId = &pb.UUID{Value: brandId}

		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}
		if image.Url != "" {
			product.Images = append(product.Images, &image)
		}

		if reviewCount.Valid {
			product.ReviewCount = reviewCount.Int32
		} else {
			product.ReviewCount = 0
		}
		if averageRating.Valid {
			product.AverageRating = float32(averageRating.Float64)
		} else {
			product.AverageRating = 0.0
		}

		if categoryName.Valid {
			product.CategoryName = categoryName.String
		} else {
			product.CategoryName = ""
		}

		if subCategoryName.Valid {
			product.SubCategoryName = subCategoryName.String
		} else {
			product.SubCategoryName = ""
		}

		if brandName.Valid {
			product.BrandName = brandName.String
		} else {
			product.BrandName = ""
		}

		if createdAt.Valid {
			product.CreatedAt = timestamppb.New(createdAt.Time)
		} else {
			product.CreatedAt = timestamppb.New(time.Time{})
		}

		if updatedAt.Valid {
			product.UpdatedAt = timestamppb.New(updatedAt.Time)
		} else {
			product.UpdatedAt = timestamppb.New(time.Time{})
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
	stmt := `UPDATE products SET name=$1, description=$2, category_id=$3, sub_category_id=$4, brand_id=$5, price=$6, quantity=$7, featured=$8 WHERE id=$9`
	product := req.Product

	_, err := s.db.Exec(stmt,
		product.Name,
		product.Description,
		product.CategoryId.Value,
		product.SubCategoryId.Value,
		product.BrandId.Value,
		product.Price,
		product.Quantity,
		product.Featured,
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

// GetProductsQueryBuilder

func BuildProductQuery(whereClause string, additionalClauses string) string {
	baseQuery := `
	WITH paginated_products AS (
		SELECT *,
		COUNT(*) OVER() AS total_count
		FROM products
		%s -- WHERE clause will be inserted here
		ORDER BY id
		LIMIT $1
		OFFSET $2
	)
	SELECT
		p.id,
		p.name,
		p.description,
		p.category_id,
		p.sub_category_id,
		p.brand_id,
		p.price,
		p.quantity,
		p.featured,
		p.slug,
		pc.name as category_name,
		psc.name as sub_category_name,
		pb.name as brand_name,
		pi.url AS image_url,
		pi.image_type,
		pr.review_count,
		pr.average_rating,
		p.created_at,
		p.updated_at,
		p.total_count
	FROM
		paginated_products p
	LEFT JOIN productimages pi ON p.id = pi.product_id
	LEFT JOIN (
		SELECT
			product_id,
			COUNT(*) AS review_count,
			AVG(rating) AS average_rating
		FROM product_reviews
		GROUP BY product_id
	) pr ON p.id = pr.product_id
	LEFT JOIN product_sub_category psc ON p.sub_category_id = psc.id
	LEFT JOIN product_category pc ON p.category_id = pc.id
	LEFT JOIN product_brand pb ON p.brand_id = pb.id
	%s; -- Additional clauses like GROUP BY or ORDER BY etc.
`
	where := ""
	if whereClause != "" {
		where = fmt.Sprintf("WHERE %s", whereClause)
	}

	return fmt.Sprintf(baseQuery, where, additionalClauses)
}

// a function that takes *sql.Rows and returns a slice of *pb.Product
func scanProducts(rows *sql.Rows) ([]*pb.Product, error) {
	defer rows.Close()

	productsMap := make(map[string]*pb.Product)

	for rows.Next() {
		var product pb.Product
		var productId, categoryId, subCategoryId, brandId string
		var categoryName, subCategoryName, brandName, imageUrl, imageType sql.NullString
		var reviewCount sql.NullInt32
		var averageRating sql.NullFloat64
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&categoryId,
			&subCategoryId,
			&brandId,
			&product.Price,
			&product.Quantity,
			&product.Featured,
			&product.Slug,
			&categoryName,
			&subCategoryName,
			&brandName,
			&imageUrl,
			&imageType,
			&reviewCount,
			&averageRating,
			&createdAt,
			&updatedAt,
			&product.TotalCount,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning products: %v",
				err,
			)
		}
		product.Id = &pb.UUID{Value: productId}
		product.CategoryId = &pb.UUID{Value: categoryId}
		product.SubCategoryId = &pb.UUID{Value: subCategoryId}
		product.BrandId = &pb.UUID{Value: brandId}

		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}

		if reviewCount.Valid {
			product.ReviewCount = reviewCount.Int32
		} else {
			product.ReviewCount = 0
		}
		if averageRating.Valid {
			product.AverageRating = float32(averageRating.Float64)
		} else {
			product.AverageRating = 0.0
		}

		if categoryName.Valid {
			product.CategoryName = categoryName.String
		} else {
			product.CategoryName = ""
		}
		if subCategoryName.Valid {
			product.SubCategoryName = subCategoryName.String
		} else {
			product.SubCategoryName = ""
		}
		if brandName.Valid {
			product.BrandName = brandName.String
		} else {
			product.BrandName = ""
		}

		if createdAt.Valid {
			product.CreatedAt = timestamppb.New(createdAt.Time)
		} else {
			product.CreatedAt = timestamppb.New(time.Time{})
		}
		if updatedAt.Valid {
			product.UpdatedAt = timestamppb.New(updatedAt.Time)
		} else {
			product.UpdatedAt = timestamppb.New(time.Time{})
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
	return products, nil
}

func getMaxPages(product *pb.Product, limit int32) int32 {
	total_count := product.GetTotalCount()
	return int32(math.Ceil(float64(total_count) / float64(limit)))
}

func (s *ProductService) GetProducts(
	ctx context.Context,
	req *pb.GetProductsRequest,
) (*pb.GetProductsResponse, error) {
	stmt := BuildProductQuery("", "ORDER BY created_at DESC")
	rows, err := s.db.Query(stmt, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}
	max_pages := getMaxPages(products[0], req.Limit)

	return &pb.GetProductsResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
		MaxPages: max_pages,
	}, nil
}

func (s *ProductService) GetProductsByCategory(
	ctx context.Context,
	req *pb.GetProductsByCategoryRequest,
) (*pb.GetProductsByCategoryResponse, error) {
	stmt := BuildProductQuery("category_id=$3", "ORDER BY created_at DESC")

	rows, err := s.db.Query(stmt, req.Limit, req.Page, req.CategoryId.Value)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error scanning products: %v",
			err,
		)
	}
	max_pages := getMaxPages(products[0], req.Limit)

	return &pb.GetProductsByCategoryResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
		MaxPages: max_pages,
	}, nil
}

func (s *ProductService) GetProductsBySubCategory(
	ctx context.Context,
	req *pb.GetProductsBySubCategoryRequest,
) (*pb.GetProductsBySubCategoryResponse, error) {
	stmt := BuildProductQuery("sub_category_id=$3", "ORDER BY created_at DESC")

	rows, err := s.db.Query(stmt, req.Limit, req.Page, req.SubCategoryId.Value)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error scanning products: %v",
			err,
		)
	}

	max_pages := getMaxPages(products[0], req.Limit)
	return &pb.GetProductsBySubCategoryResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
		MaxPages: max_pages,
	}, nil
}

func (s *ProductService) GetProductsByBrand(
	ctx context.Context,
	req *pb.GetProductsByBrandRequest,
) (*pb.GetProductsByBrandResponse, error) {
	stmt := BuildProductQuery("brand_id=$3", "ORDER BY created_at DESC")
	rows, err := s.db.Query(stmt, req.Limit, req.Page, req.BrandId.Value)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error scanning products: %v",
			err,
		)
	}

	max_pages := getMaxPages(products[0], req.Limit)
	return &pb.GetProductsByBrandResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
		MaxPages: max_pages,
	}, nil
}

func (s *ProductService) GetFeaturedProducts(ctx context.Context, req *pb.GetFeaturedProductsRequest) (*pb.GetFeaturedProductsResponse, error) {
	stmt := BuildProductQuery("featured=true", "ORDER BY created_at DESC")

	rows, err := s.db.Query(stmt, req.Limit, req.Page)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting products: %v",
			err,
		)
	}

	defer rows.Close()

	products, err := scanProducts(rows)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error scanning products: %v",
			err,
		)
	}

	max_pages := getMaxPages(products[0], req.Limit)
	return &pb.GetFeaturedProductsResponse{
		Products: products,
		Limit:    req.Limit,
		Page:     req.Page,
		MaxPages: max_pages,
	}, nil
}

func (s *ProductService) GetProductBySlug(ctx context.Context, req *pb.GetProductBySlugRequest) (*pb.GetProductBySlugResponse, error) {
	stmt := `
SELECT
  p.id,
  p.name,
  p.description,
  p.category_id,
  p.sub_category_id,
  p.brand_id,
  p.price,
  p.quantity,
  p.featured,
  p.slug,
  pc.name AS category_name,
  psc.name AS sub_category_name,
  pb.name AS brand_name,
  pi.url AS image_url,
  pi.image_type,
  pr.review_count,
  pr.average_rating,
  p.created_at,
  p.updated_at
FROM
  products p
LEFT JOIN (
  SELECT
    product_id,
    COUNT(*) AS review_count,
    AVG(rating) AS average_rating
  FROM
    product_reviews
  GROUP BY
    product_id
) pr ON p.id = pr.product_id
LEFT JOIN
  productimages pi ON p.id = pi.product_id
LEFT JOIN
 product_category pc ON p.category_id = pc.id
LEFT JOIN
  product_sub_category psc ON p.sub_category_id = psc.id
LEFT JOIN
  product_brand pb ON p.brand_id = pb.id
WHERE
  p.slug = $1;`

	decodedSlug, err := url.QueryUnescape(req.Slug)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"error decoding slug: %+v",
			err,
		)
	}

	rows, err := s.db.Query(stmt, decodedSlug)
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
		var reviewCount sql.NullInt32
		var averageRating sql.NullFloat64
		var categoryName, subCategoryName, brandName sql.NullString
		var createdAt, updatedAt sql.NullTime
		var brandId, categoryId, subCategoryId string

		err := rows.Scan(
			&productId,
			&product.Name,
			&product.Description,
			&categoryId,
			&subCategoryId,
			&brandId,
			&product.Price,
			&product.Quantity,
			&product.Featured,
			&product.Slug,
			&categoryName,
			&subCategoryName,
			&brandName,
			&imageUrl,
			&imageType,
			&reviewCount,
			&averageRating,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning rows: %v",
				err.Error(),
			)
		}
		product.CategoryId = &pb.UUID{Value: categoryId}
		product.SubCategoryId = &pb.UUID{Value: subCategoryId}
		product.BrandId = &pb.UUID{Value: brandId}

		image := pb.Image{
			Url:       imageUrl.String,
			ImageType: imageType.String,
		}
		if image.Url != "" {
			product.Images = append(product.Images, &image)
		}

		if reviewCount.Valid {
			product.ReviewCount = reviewCount.Int32
		} else {
			product.ReviewCount = 0
		}
		if averageRating.Valid {
			product.AverageRating = float32(averageRating.Float64)
		} else {
			product.AverageRating = 0.0
		}

		if categoryName.Valid {
			product.CategoryName = categoryName.String
		} else {
			product.CategoryName = ""
		}

		if subCategoryName.Valid {
			product.SubCategoryName = subCategoryName.String
		} else {
			product.SubCategoryName = ""
		}

		if brandName.Valid {
			product.BrandName = brandName.String
		} else {
			product.BrandName = ""
		}

		if createdAt.Valid {
			product.CreatedAt = timestamppb.New(createdAt.Time)
		} else {
			product.CreatedAt = timestamppb.New(time.Time{})
		}

		if updatedAt.Valid {
			product.UpdatedAt = timestamppb.New(updatedAt.Time)
		} else {
			product.UpdatedAt = timestamppb.New(time.Time{})
		}

		product.Id = &pb.UUID{Value: productId}

	}
	if productId == "" {
		return nil, status.Errorf(
			codes.NotFound,
			"product with id %s not found.",
			req.Slug,
		)
	}
	return &pb.GetProductBySlugResponse{
		Product: &product,
	}, nil
}

func (p *ProductService) GetProductsByCategorySlug(ctx context.Context, req *pb.GetProductsByCategorySlugRequest) (*pb.GetProductsByCategorySlugResponse, error) {
	// first get the id of the slug.
	resp, err := p.GetCategoryBySlug(ctx, &pb.GetCategoryBySlugRequest{Slug: req.CategorySlug})
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"category with slug %s not found.",
			req.CategorySlug,
		)
	}

	categoryId := resp.Category.Id

	// now get the products by category id.
	productResp, err := p.GetProductsByCategory(ctx, &pb.GetProductsByCategoryRequest{CategoryId: categoryId, Limit: req.Limit, Page: req.Page})
	if err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			"products with category id %s not found.",
			categoryId,
		)
	}

	return &pb.GetProductsByCategorySlugResponse{
		Products: productResp.Products,
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
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"could not upload the image %s",
			err.Error(),
		)
	}

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

// Reviews

func (s *ProductService) CreateReview(ctx context.Context, request *pb.CreateReviewRequest) (*pb.CreateReviewResponse, error) {
	stmt := `INSERT INTO product_reviews (product_id, user_id, rating, title, content) VALUES ($1, $2, $3, $4, $5)`
	// convert the Ids from id.value to the required uuid format required by the database
	_, err := s.db.Exec(stmt, request.ProductId.Value, request.UserId.Value, request.Rating, request.Title, request.Content)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error creating review: %v",
			err.Error(),
		)
	}

	return &pb.CreateReviewResponse{
		Message: "Review created successfully",
	}, nil
}

func (s *ProductService) GetReviews(ctx context.Context, request *pb.GetReviewsRequest) (*pb.GetReviewsResponse, error) {
	// left join with user table
	stmt := `SELECT r.id, r.product_id, r.user_id, r.rating, r.title, r.content, u.name as user_name
	FROM product_reviews r
	LEFT JOIN users u ON r.user_id = u.id
	WHERE r.product_id = $1;`

	rows, err := s.db.Query(stmt, request.ProductId.Value)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting reviews: %v",
			err.Error(),
		)
	}

	var reviews []*pb.Review

	for rows.Next() {
		var review pb.Review
		var reviewIdStr string
		var reviewProductIdStr string
		var reviewUserIdStr string

		err := rows.Scan(&reviewIdStr, &reviewProductIdStr, &reviewUserIdStr, &review.Rating, &review.Title, &review.Content, &review.UserName)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning review %v",
				err.Error(),
			)
		}
		review.Id = &pb.UUID{
			Value: reviewIdStr,
		}
		review.ProductId = &pb.UUID{
			Value: reviewProductIdStr,
		}
		review.UserId = &pb.UUID{
			Value: reviewUserIdStr,
		}

		reviews = append(reviews, &review)
	}

	return &pb.GetReviewsResponse{
		Reviews: reviews,
	}, nil
}

func (s *ProductService) GetReview(ctx context.Context, request *pb.GetReviewRequest) (*pb.GetReviewResponse, error) {
	stmt := `SELECT r.id, r.product_id, r.user_id, r.rating, r.title, r.content, u.name FROM product_reviews r JOIN users u ON r.user_id = u.id WHERE r.id=$1`
	row := s.db.QueryRow(stmt, request.ReviewId.Value)

	var review pb.Review
	var reviewIdStr string
	var productIdStr string
	var userIdStr string
	var userName string
	err := row.Scan(&reviewIdStr, &productIdStr, &userIdStr, &review.Rating, &review.Title, &review.Content, &userName)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting review: %v",
			err.Error(),
		)
	}

	review.Id = &pb.UUID{
		Value: reviewIdStr,
	}
	review.ProductId = &pb.UUID{
		Value: productIdStr,
	}
	review.UserId = &pb.UUID{
		Value: userIdStr,
	}
	review.UserName = userName

	return &pb.GetReviewResponse{
		Review: &review,
	}, nil
}

func (s *ProductService) GetProductRating(ctx context.Context, request *pb.GetProductRatingRequest) (*pb.GetProductRatingResponse, error) {
	stmt := `SELECT AVG(rating), COUNT(*) FROM product_reviews WHERE product_id=$1`

	log.Printf("The productId is: %s ", request.ProductId.Value)

	row := s.db.QueryRow(stmt, request.ProductId.Value)

	var rating float64
	var count int64
	err := row.Scan(&rating, &count)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting product rating: %v",
			err.Error(),
		)
	}

	return &pb.GetProductRatingResponse{
		AverageRating:   float32(rating),
		NumberOfReviews: int32(count),
	}, nil
}

func (s *ProductService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	stmt := `INSERT INTO product_category(name, description, featured) VALUES ($1, $2, $3) RETURNING id;`
	row := s.db.QueryRow(stmt, req.Name, req.Description, req.Featured)

	var id string
	err := row.Scan(&id)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error creating category: %v",
			err.Error(),
		)
	}

	return &pb.CreateCategoryResponse{
		Message: "category created successfully",
	}, nil
}

func (s *ProductService) GetCategories(ctx context.Context, req *pb.GetCategoriesRequest) (*pb.GetCategoriesResponse, error) {
	stmt := `SELECT id, name, description, featured, slug FROM product_category LIMIT $1 OFFSET $2`
	rows, err := s.db.Query(stmt, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting categories: %v",
			err.Error(),
		)
	}
	defer rows.Close()

	var categories []*pb.Category
	for rows.Next() {
		var categoryId string
		var category pb.Category
		err := rows.Scan(&categoryId, &category.Name, &category.Description, &category.Featured, &category.Slug)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning category: %v",
				err.Error(),
			)
		}
		category.Id = &pb.UUID{Value: categoryId}
		categories = append(categories, &category)
	}

	return &pb.GetCategoriesResponse{
		Categories: categories,
	}, nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	stmt := `UPDATE product_category SET name = $1, description = $2, featured = $3 WHERE id = $4`
	_, err := s.db.Exec(stmt, req.Name, req.Description, req.Featured, req.Id.Value)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error updating category: %v",
			err.Error(),
		)
	}

	return &pb.UpdateCategoryResponse{
		Message: "updated category successfully",
	}, nil
}

func (s *ProductService) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	stmt := `DELETE FROM product_category WHERE id = $1`
	_, err := s.db.Exec(stmt, req.Id.Value)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error deleting category: %v",
			err.Error(),
		)
	}

	return &pb.DeleteCategoryResponse{
		Message: "deleted category successfully",
	}, nil
}

func (s *ProductService) GetFeaturedCategories(ctx context.Context, req *pb.GetFeaturedCategoriesRequest) (*pb.GetFeaturedCategoriesResponse, error) {
	stmt := `SELECT id, name, description, featured, slug FROM product_category WHERE featured = true`
	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting featured categories: %v",
			err.Error(),
		)
	}
	defer rows.Close()

	var categories []*pb.Category
	for rows.Next() {
		var category pb.Category
		var categoryId string
		err := rows.Scan(&categoryId, &category.Name, &category.Description, &category.Featured, &category.Slug)
		if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning category: %v",
				err.Error(),
			)
		}
		category.Id = &pb.UUID{Value: categoryId}
		categories = append(categories, &category)
	}

	return &pb.GetFeaturedCategoriesResponse{
		Categories: categories,
	}, nil
}

func (s *ProductService) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	stmt := `SELECT id, name, description, slug, featured FROM product_category WHERE id = $1`
	row := s.db.QueryRow(stmt, req.Id.Value)
	var category pb.Category
	var categoryId string
	err := row.Scan(&categoryId, &category.Name, &category.Description, &category.Slug, &category.Featured)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting category: %v",
			err.Error(),
		)
	}

	category.Id = &pb.UUID{Value: categoryId}
	return &pb.GetCategoryResponse{
		Category: &category,
	}, nil
}

func (s *ProductService) GetCategoryBySlug(ctx context.Context, req *pb.GetCategoryBySlugRequest) (*pb.GetCategoryResponse, error) {
	stmt := `SELECT id, name, description, slug, featured FROM product_category WHERE slug = $1`
	decodedSlug, err := url.QueryUnescape(req.Slug)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid slug: %v",
			err.Error(),
		)
	}
	row := s.db.QueryRow(stmt, decodedSlug)
	var category pb.Category
	var categoryId string
	err = row.Scan(&categoryId, &category.Name, &category.Description, &category.Slug, &category.Featured)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting category: %v",
			err.Error(),
		)
	}

	category.Id = &pb.UUID{Value: categoryId}
	return &pb.GetCategoryResponse{
		Category: &category,
	}, nil
}

func (s *ProductService) CreateSubCategory(ctx context.Context, req *pb.CreateSubCategoryRequest) (*pb.CreateSubCategoryResponse, error) {
	stmt := `INSERT INTO product_sub_category (name, description, category_id, slug) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(stmt, req.Name, req.Description, req.CategoryId.Value, req.Slug)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error creating subcategory: %v",
			err.Error(),
		)
	}

	return &pb.CreateSubCategoryResponse{
		Message: "Subcategory created successfully",
	}, nil
}

func (s *ProductService) GetSubCategories(ctx context.Context, req *pb.GetSubCategoriesRequest) (*pb.GetSubCategoriesResponse, error) {
	stmt := `SELECT id, name, description, category_id, slug FROM product_sub_category WHERE category_id = $1 LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(stmt, req.CategoryId.Value, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting subcategories: %v",
			err.Error(),
		)
	}
	defer rows.Close()

	var subCategories []*pb.SubCategory
	var subCategoryId, categoryId string
	for rows.Next() {
		var subCategory pb.SubCategory
		if err := rows.Scan(&subCategoryId, &subCategory.Name, &subCategory.Description, &categoryId, &subCategory.Slug); err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning subcategory: %v",
				err.Error(),
			)
		}

		subCategory.Id = &pb.UUID{Value: subCategoryId}
		subCategory.CategoryId = &pb.UUID{Value: categoryId}
		subCategories = append(subCategories, &subCategory)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error iterating over subcategories: %v",
			err.Error(),
		)
	}

	return &pb.GetSubCategoriesResponse{
		SubCategories: subCategories,
	}, nil
}

func (s *ProductService) UpdateSubCategory(ctx context.Context, req *pb.UpdateSubCategoryRequest) (*pb.UpdateSubCategoryResponse, error) {
	stmt := `UPDATE product_sub_category SET name = $1, description = $2, slug = $3 WHERE id = $4`
	if _, err := s.db.Exec(stmt, req.Name, req.Description, req.Slug, req.Id.Value); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error updating subcategory: %v",
			err.Error(),
		)
	}

	return &pb.UpdateSubCategoryResponse{
		Message: "updated subcategory successfully",
	}, nil
}

func (s *ProductService) DeleteSubCategory(ctx context.Context, req *pb.DeleteSubCategoryRequest) (*pb.DeleteSubCategoryResponse, error) {
	stmt := `DELETE FROM product_sub_category WHERE id = $1`
	if _, err := s.db.Exec(stmt, req.Id.Value); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error deleting subcategory: %v",
			err.Error(),
		)
	}

	return &pb.DeleteSubCategoryResponse{
		Message: "deleted subcategory successfully",
	}, nil
}

func (s *ProductService) GetSubCategory(ctx context.Context, req *pb.GetSubCategoryRequest) (*pb.GetSubCategoryResponse, error) {
	stmt := `SELECT id, name, description, category_id, slug FROM product_sub_category WHERE id = $1`
	row := s.db.QueryRow(stmt, req.Id.Value)
	var subCategory pb.SubCategory
	var subCategoryId string
	var categoryId string
	if err := row.Scan(&subCategoryId, &subCategory.Name, &subCategory.Description, &categoryId, &subCategory.Slug); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error scanning subcategory: %v",
			err.Error(),
		)
	}

	subCategory.Id = &pb.UUID{Value: subCategoryId}
	subCategory.CategoryId = &pb.UUID{Value: categoryId}

	return &pb.GetSubCategoryResponse{
		SubCategory: &subCategory,
	}, nil
}

func (p *ProductService) GetSubCategoryBySlug(ctx context.Context, req *pb.GetSubCategoryBySlugRequest) (*pb.GetSubCategoryResponse, error) {
	stmt := `SELECT id, name, description, category_id, slug FROM product_sub_category WHERE slug = $1`
	decodedSlug, err := url.QueryUnescape(req.Slug)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"error decoding slug: %v",
			err.Error(),
		)
	}
	row := p.db.QueryRow(stmt, decodedSlug)
	var subCategory pb.SubCategory
	var subCategoryId string
	var categoryId string
	if err := row.Scan(&subCategoryId, &subCategory.Name, &subCategory.Description, &categoryId, &subCategory.Slug); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error scanning subcategory: %v",
			err.Error(),
		)
	}

	subCategory.Id = &pb.UUID{Value: subCategoryId}
	subCategory.CategoryId = &pb.UUID{Value: categoryId}

	return &pb.GetSubCategoryResponse{
		SubCategory: &subCategory,
	}, nil
}

func (s *ProductService) CreateBrand(ctx context.Context, req *pb.CreateBrandRequest) (*pb.CreateBrandResponse, error) {
	stmt := `INSERT INTO product_brand (name, description) VALUES ($1, $2) RETURNING id`

	var id string
	err := s.db.QueryRow(stmt, req.Name, req.Description).Scan(&id)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error creating brand: %v",
			err.Error(),
		)
	}

	return &pb.CreateBrandResponse{
		Message: fmt.Sprintf("Brand created successfully with ID %s", id),
	}, nil
}

func (s *ProductService) GetBrands(ctx context.Context, req *pb.GetBrandsRequest) (*pb.GetBrandsResponse, error) {
	stmt := `SELECT id, name, description FROM product_brand`
	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error getting brands: %v",
			err.Error(),
		)
	}
	defer rows.Close()

	var brands []*pb.Brand
	for rows.Next() {
		var brand pb.Brand
		var brandId string
		if err := rows.Scan(&brandId, &brand.Name, &brand.Description); err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"error scanning brand: %v",
				err.Error(),
			)
		}
		brand.Id = &pb.UUID{Value: brandId}
		brands = append(brands, &brand)
	}

	return &pb.GetBrandsResponse{
		Brands: brands,
	}, nil
}

func (s *ProductService) UpdateBrand(ctx context.Context, req *pb.UpdateBrandRequest) (*pb.UpdateBrandResponse, error) {
	stmt := `UPDATE product_brand SET name = $1, description = $2 WHERE id = $3`
	if _, err := s.db.Exec(stmt, req.Name, req.Description, req.Id.Value); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error updating brand: %v",
			err.Error(),
		)
	}

	return &pb.UpdateBrandResponse{
		Message: fmt.Sprintf("Brand updated successfully with ID %s", req.Id.Value),
	}, nil
}

func (s *ProductService) DeleteBrand(ctx context.Context, req *pb.DeleteBrandRequest) (*pb.DeleteBrandResponse, error) {
	stmt := `DELETE FROM product_brand WHERE id = $1`
	if _, err := s.db.Exec(stmt, req.Id.Value); err != nil {
		return nil, status.Errorf(
			codes.Internal,
			"error deleting brand: %v",
			err.Error(),
		)
	}

	return &pb.DeleteBrandResponse{
		Message: fmt.Sprintf("Brand deleted successfully with ID %s", req.Id.Value),
	}, nil
}

func (s *ProductService) GetBrand(ctx context.Context, req *pb.GetBrandRequest) (*pb.GetBrandResponse, error) {
	stmt := `SELECT id, name, description FROM product_brand WHERE id = $1`
	var brand pb.Brand
	var brandId string
	if err := s.db.QueryRow(stmt, req.Id.Value).Scan(&brandId, &brand.Name, &brand.Description); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "brand not found")
		}
		return nil, status.Errorf(
			codes.Internal,
			"error getting brand: %v",
			err.Error(),
		)
	}

	brand.Id = &pb.UUID{Value: brandId}
	return &pb.GetBrandResponse{
		Brand: &brand,
	}, nil
}
