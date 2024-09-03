package cmsservice

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kelcheone/chemistke/internal/database"
	"github.com/kelcheone/chemistke/pkg/codes"
	pb "github.com/kelcheone/chemistke/pkg/grpc/cms"
	"github.com/kelcheone/chemistke/pkg/status"
)

type CmsService struct {
	pb.UnimplementedCmsServiceServer
	db database.DB
}

func NewCmsService(db database.DB) *CmsService {
	return &CmsService{db: db}
}

func (c *CmsService) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	stmt := `INSERT INTO content (
  published_date,
  updated_date,
  cover_image,
  title,
  description,
  slug,
  content,
  status,
  author_id,
  category_id
  )VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	post := req.Post
	row := c.db.QueryRow(
		stmt,
		post.PublishedDate,
		post.UpdatedDate,
		post.CoverImage,
		post.Title,
		post.Description,
		post.Slug,
		post.Content,
		post.Status,
		post.AuthorId.Value,
		post.CategoryId.Value,
	)
	var postId string
	if err := row.Scan(&postId); err != nil {
		return nil, status.Errorf(codes.Internal, "could not get post Id: %v", err)
	}

	return &pb.CreatePostResponse{
		PostId: &pb.UUID{Value: postId},
	}, nil
}

func (c *CmsService) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	stmt := `SELECT id, published_date, updated_date, cover_image, title, description, slug, content, status, author_id, category_id FROM content WHERE id=$1`
	// stmt := `SELECT * FROM content WHERE id=$1`
	var post pb.Post
	var postId, categoryId, authorId string
	fmt.Println(req.PostId.Value)
	fmt.Println(stmt)
	err := c.db.QueryRow(stmt, req.PostId.Value).Scan(
		&postId,
		&post.PublishedDate,
		&post.UpdatedDate,
		&post.CoverImage,
		&post.Title,
		&post.Description,
		&post.Slug,
		&post.Content,
		&post.Status,
		&authorId,
		&categoryId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "post with id %s not found", req.PostId.Value)
		}
		return nil, status.Errorf(codes.Internal, "error getting post: %v", err)
	}

	post.PostId = &pb.UUID{Value: postId}
	post.AuthorId = &pb.UUID{Value: authorId}
	post.CategoryId = &pb.UUID{Value: categoryId}
	return &pb.GetPostResponse{Post: &post}, nil
}

func (c *CmsService) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.UpdatePostResponse, error) {
	stmt := `UPDATE content SET published_date=$1, updated_date=$2, cover_image=$3, title=$4, slug=$5, content=$6, status=$7, author_id=$8, category_id=$9, description=$10 WHERE id=$11 RETURNING id`
	post := req.Post
	_, err := c.db.Exec(
		stmt,
		post.PublishedDate,
		post.UpdatedDate,
		post.CoverImage,
		post.Title,
		post.Slug,
		post.Content,
		post.Status,
		post.AuthorId.Value,
		post.CategoryId.Value,
		post.Description,
		req.PostId.Value,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "post does not exist")
		}
		return nil, status.Errorf(codes.Internal, "could not update post: %v", err)
	}

	return &pb.UpdatePostResponse{PostId: post.PostId}, nil
}

func (c *CmsService) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	stmt := `DELETE FROM content WHERE id=$1`
	_, err := c.db.Exec(stmt, req.PostId.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "post does not exist")
		}
		return nil, status.Errorf(codes.Internal, "could not delet post ")
	}
	return nil, nil
}

func (c *CmsService) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	stmt := `SELECT id, published_date, updated_date, cover_image, title, description,
  slug, content, status, author_id, category_id FROM content LIMIT $1 OFFSET $2`

	rows, err := c.db.Query(stmt, req.PerPage, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query not successful: %v", err.Error())
	}

	var posts []*pb.Post
	for rows.Next() {

		post, err := PostRowScanner(rows)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)

	}

	return &pb.ListPostsResponse{Posts: posts}, nil
}

func (c *CmsService) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.CreateCategoryResponse, error) {
	stmt := `INSERT INTO categories (name, slug, description) VALUES($1, $2, $3) RETURNING id`
	row := c.db.QueryRow(stmt, req.Category.Name, req.Category.Slug, req.Category.Description)
	var categoryId string
	err := row.Scan(&categoryId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not get the category id: %v", err)
	}
	return &pb.CreateCategoryResponse{CategoryId: &pb.UUID{Value: categoryId}}, nil
}

func (c *CmsService) GetCategory(ctx context.Context, req *pb.GetCategoryRequest) (*pb.GetCategoryResponse, error) {
	stmt := `SELECT id, name, slug, description FROM categories WHERE id=$1`
	row := c.db.QueryRow(stmt, req.CategoryId.Value)
	var category pb.Category
	var categoryId string
	err := row.Scan(
		&categoryId,
		&category.Name,
		&category.Slug,
		&category.Description,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "category with id %s not found", req.CategoryId.Value)
		}
		return nil, status.Errorf(codes.Internal, "could not get category: %v", err.Error())
	}
	category.CategoryId = &pb.UUID{Value: categoryId}
	return &pb.GetCategoryResponse{Category: &category}, nil
}

func (c *CmsService) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.UpdateCategoryResponse, error) {
	stmt := `UPDATE categories name=$1, slug=$2, description=$3 WHERE id=$4`
	_, err := c.db.Exec(
		stmt,
		req.Category.Name,
		req.Category.Slug,
		req.Category.Description,
		req.Category.CategoryId.Value,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "category does not exist")
		}
		return nil, status.Errorf(codes.Internal, "could not update category: %v", err.Error())
	}
	return nil, nil
}

func (c *CmsService) DeleteCategory(ctx context.Context, req *pb.DeleteCategoryRequest) (*pb.DeleteCategoryResponse, error) {
	stmt := `DELETE FROM categories WHERE id=$1`
	_, err := c.db.Exec(stmt, req.CategoryId.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "category does not exist")
		}
		return nil, status.Errorf(codes.Internal, "could not delete category: %v", err.Error())
	}
	return &pb.DeleteCategoryResponse{CategoryId: req.CategoryId}, nil
}

func (c *CmsService) ListCategories(ctx context.Context, req *pb.ListCategoriesRequest) (*pb.ListCategoriesResponse, error) {
	stmt := `SELECT id, name, slug, description FROM categories LIMIT $1 OFFSET $2`

	rows, err := c.db.Query(stmt, req.PerPage, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not fetch categories: %v", err.Error())
	}

	var categories []*pb.Category
	for rows.Next() {
		category, err := CategoryRowScanner(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return &pb.ListCategoriesResponse{Categories: categories}, nil
}

func (c *CmsService) GetCategoryPosts(ctx context.Context, req *pb.GetCategoryPostsRequest) (*pb.GetCategoryPostsResponse, error) {
	stmt := `SELECT id, published_date, updated_date, cover_image, title, description,
  slug, content, status, author_id, category_id FROM content WHERE category_id=$1 LIMIT $2 OFFSET $3 `

	rows, err := c.db.Query(stmt, req.CategoryId.Value, req.PerPage, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query not successful: %v", err.Error())
	}

	var posts []*pb.Post
	for rows.Next() {
		post, err := PostRowScanner(rows)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}
	return &pb.GetCategoryPostsResponse{Posts: posts}, nil
}

func (c *CmsService) CreateAuthor(ctx context.Context, req *pb.CreateAuthorRequest) (*pb.CreateAuthorResponse, error) {
	stmt := `INSERT INTO authors (bio, avatar, url, user_id) VALUES ($1, $2, $3, $4) RETURNING id`
	author := req.Author
	row := c.db.QueryRow(stmt, author.Bio, author.Avatar, author.Url, author.UserId.Value)
	var authorId string
	err := row.Scan(&authorId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not create author: %v", err.Error())
	}

	return &pb.CreateAuthorResponse{AuthorId: &pb.UUID{Value: authorId}}, nil
}

func (c *CmsService) GetAuthor(ctx context.Context, req *pb.GetAuthorRequest) (*pb.GetAuthorResponse, error) {
	stmt := `SELECT id, bio, avatar, url, user_id FROM authors WHERE id=$1`
	var author pb.Author
	var authorId, userId string
	row := c.db.QueryRow(stmt, req.AuthorId.Value)
	err := row.Scan(
		&authorId,
		&author.Bio,
		&author.Avatar,
		&author.Url,
		&userId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "could not find author with id %s", req.AuthorId.Value)
		}
		return nil, status.Errorf(codes.Internal, "could not get author: %v", err.Error())
	}
	author.AuthorId = &pb.UUID{Value: authorId}
	author.UserId = &pb.UUID{Value: userId}
	return &pb.GetAuthorResponse{Author: &author}, nil
}

func (c *CmsService) UpdateAuthor(ctx context.Context, req *pb.UpdateAuthorRequest) (*pb.UpdateAuthorResponse, error) {
	stmt := `UPDATE authors SET bio=$1, avatar=$2, url=$3 WHERE id=$4`
	_, err := c.db.Exec(stmt, req.Author.Bio, req.Author.Avatar, req.Author.Url, req.AuthorId.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "author does not exist")
		}
		return nil, status.Errorf(codes.Internal, "could not update author: %v", err.Error())
	}
	return &pb.UpdateAuthorResponse{AuthorId: req.AuthorId}, nil
}

func (c *CmsService) DeleteAuthor(ctx context.Context, req *pb.DeleteAuthorRequest) (*pb.DeleteAuthorResponse, error) {
	stmt := `DELETE FROM authors WHERE id=$1`

	_, err := c.db.Exec(stmt, req.AuthorId.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "author does not exist")
		}
		return nil, status.Errorf(codes.Internal, "could not delete author: %v", err)
	}
	return &pb.DeleteAuthorResponse{AuthorId: req.AuthorId}, nil
}

func (c *CmsService) ListAuthors(ctx context.Context, req *pb.ListAuthorsRequest) (*pb.ListAuthorsResponse, error) {
	stmt := `SELECT id, bio, avatar, url, user_id FROM authors LIMIT $1 OFFSET $2`
	rows, err := c.db.Query(stmt, req.PerPage, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could get authors: %v", err.Error())
	}

	var authors []*pb.Author
	for rows.Next() {
		var author pb.Author
		var authorId, userId string
		err := rows.Scan(
			&authorId,
			&author.Bio,
			&author.Avatar,
			&author.Url,
			&userId,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "could not get author: %v", err.Error())
		}
		author.AuthorId = &pb.UUID{Value: authorId}
		author.UserId = &pb.UUID{Value: userId}

		authors = append(authors, &author)
	}
	return &pb.ListAuthorsResponse{Authors: authors}, nil
}

func (c *CmsService) GetAuthorPosts(ctx context.Context, req *pb.GetAuthorPostsRequest) (*pb.GetAuthorPostsResponse, error) {
	stmt := `SELECT id, published_date, updated_date, cover_image, title, description,
  slug, content, status, author_id, category_id FROM content WHERE author_id=$1 LIMIT $2 OFFSET $3 `

	rows, err := c.db.Query(stmt, req.AuthorId.Value, req.PerPage, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query not successful: %v", err.Error())
	}

	var posts []*pb.Post
	for rows.Next() {
		post, err := PostRowScanner(rows)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}
	return &pb.GetAuthorPostsResponse{Posts: posts}, nil
}

func (c *CmsService) GetAuthorCategoryPosts(ctx context.Context, req *pb.GetAuthorCategoryPostsRequest) (*pb.GetAuthorCategoryPostsResponse, error) {
	stmt := `SELECT id, published_date, updated_date, cover_image, title, description, url,
  slug, content, status, author_id, category_id FROM content WHERE category_id=$1 AND author_id=$2 LIMIT $3 OFFSET $4 `

	rows, err := c.db.Query(stmt, req.CategoryId.Value, req.AuthorId.Value, req.PerPage, req.Page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query not successful: %v", err.Error())
	}

	var posts []*pb.Post
	for rows.Next() {
		post, err := PostRowScanner(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return &pb.GetAuthorCategoryPostsResponse{Posts: posts}, nil
}

func PostRowScanner(rows *sql.Rows) (*pb.Post, error) {
	var post pb.Post
	var postId, categoryId, authorId string

	err := rows.Scan(
		&postId,
		&post.PublishedDate,
		&post.UpdatedDate,
		&post.CoverImage,
		&post.Title,
		&post.Description,
		&post.Slug,
		&post.Content,
		&post.Status,
		&authorId,
		&categoryId,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not scan post: %v", err)
	}

	post.PostId = &pb.UUID{Value: postId}
	post.AuthorId = &pb.UUID{Value: authorId}
	post.CategoryId = &pb.UUID{Value: categoryId}

	return &post, nil
}

func CategoryRowScanner(rows *sql.Rows) (*pb.Category, error) {
	var category pb.Category
	var categoryId string
	err := rows.Scan(
		&categoryId,
		&category.Name,
		&category.Slug,
		&category.Description,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not scan category: %v", err)
	}
	category.CategoryId = &pb.UUID{Value: categoryId}
	return &category, nil
}
