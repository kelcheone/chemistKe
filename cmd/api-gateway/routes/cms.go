package routes

import (
	"fmt"
	"net/http"

	"github.com/kelcheone/chemistke/cmd/utils"
	cms_proto "github.com/kelcheone/chemistke/pkg/grpc/cms"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Post struct {
	Id            string `json:"id"`
	PublishedDate string `json:"published_date"`
	UpdatedDate   string `json:"updated_date"`
	CoverImage    string `json:"cover_image"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Slug          string `json:"slug"`
	Content       string `json:"content"`
	Status        string `json:"status"`
	AuthorId      string `json:"author_id"`
	CategroyId    string `json:"category_id"`
}

type Author struct {
	Id     string `json:"id"`
	Bio    string `json:"bio"`
	Avatar string `json:"avatar"`
	Url    string `json:"url"`
	UserId string `json:"user_id"`
}

type Category struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type CmsServer struct {
	CmsClient cms_proto.CmsServiceClient
}

func ConnectCmsServer(link string) (*CmsServer, func(), error) {
	cmsConn, err := grpc.NewClient(
		link,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not connect to order server: %v",
			err,
		)
	}

	conn := &CmsServer{
		CmsClient: cms_proto.NewCmsServiceClient(cmsConn),
	}

	return conn, func() {
		cmsConn.Close()
	}, nil
}

func (s *CmsServer) CreateAuthor(c echo.Context) error {
	var author Author

	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	// here we should update the user Role to Author.
	if err := c.Bind(&author); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	if author.UserId == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "user id is required",
		})
	}

	_, err := s.CmsClient.UpdateUserRole(
		c.Request().Context(),
		&cms_proto.UpdateUserRoleRequest{
			UserId: &cms_proto.UUID{Value: author.UserId},
			Role:   cms_proto.UserRoles_AUTHOR,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	resp, err := s.CmsClient.CreateAuthor(
		c.Request().Context(),
		&cms_proto.CreateAuthorRequest{
			Author: &cms_proto.Author{
				Bio:    author.Bio,
				Url:    author.Url,
				Avatar: author.Avatar,
				UserId: &cms_proto.UUID{Value: author.UserId},
			},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) GetAuthor(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id was not provided",
		})
	}

	resp, err := s.CmsClient.GetAuthor(
		c.Request().Context(),
		&cms_proto.GetAuthorRequest{
			AuthorId: &cms_proto.UUID{Value: id},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) UpdateAuthor(c echo.Context) error {
	var author Author

	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin && !claims.Author {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	if err := c.Bind(&author); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	if author.Id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "author id was not provided",
		})
	}

	resp, err := s.CmsClient.UpdateAuthor(
		c.Request().Context(),
		&cms_proto.UpdateAuthorRequest{
			AuthorId: &cms_proto.UUID{Value: author.Id},
			Author: &cms_proto.Author{
				AuthorId: &cms_proto.UUID{Value: author.Id},
				Bio:      author.Bio,
				Avatar:   author.Avatar,
				UserId:   &cms_proto.UUID{Value: author.UserId},
				Url:      author.Url,
			},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, resp)
}

func (s *CmsServer) DeleteAuthor(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id was not provided",
		})
	}

	resp, err := s.CmsClient.DeleteAuthor(
		c.Request().Context(),
		&cms_proto.DeleteAuthorRequest{
			AuthorId: &cms_proto.UUID{Value: id},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) ListAuthors(c echo.Context) error {
	var req PaginatedReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}
	resp, err := s.CmsClient.ListAuthors(
		c.Request().Context(),
		&cms_proto.ListAuthorsRequest{
			Page:    req.Page,
			PerPage: req.Limit,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) CreateCategory(c echo.Context) error {
	var category Category

	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin && !claims.Author {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	if err := c.Bind(&category); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	nReq := &cms_proto.CreateCategoryRequest{
		Category: &cms_proto.Category{
			Description: category.Description,
			Name:        category.Name,
			Slug:        category.Slug,
		},
	}

	resp, err := s.CmsClient.CreateCategory(c.Request().Context(), nReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) GetCategory(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id was not provided",
		})
	}

	resp, err := s.CmsClient.GetCategory(
		c.Request().Context(),
		&cms_proto.GetCategoryRequest{CategoryId: &cms_proto.UUID{Value: id}},
	)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: err.Error()},
		)
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) UpdateCategory(c echo.Context) error {
	var category Category

	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin && !claims.Author {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	if err := c.Bind(&category); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	if category.Id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "cateogory Id was not provided",
		})
	}

	nReq := &cms_proto.UpdateCategoryRequest{
		CategoryId: &cms_proto.UUID{Value: category.Id},
		Category: &cms_proto.Category{
			CategoryId:  &cms_proto.UUID{Value: category.Id},
			Name:        category.Name,
			Slug:        category.Slug,
			Description: category.Description,
		},
	}

	resp, err := s.CmsClient.UpdateCategory(c.Request().Context(), nReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, resp)
}

func (s *CmsServer) DeleteCategory(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin && !claims.Author {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id was not provided",
		})
	}

	resp, err := s.CmsClient.DeleteCategory(
		c.Request().Context(),
		&cms_proto.DeleteCategoryRequest{
			CategoryId: &cms_proto.UUID{Value: id},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusNoContent, resp)
}

func (s *CmsServer) ListCategories(c echo.Context) error {
	var req PaginatedReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	resp, err := s.CmsClient.ListCategories(
		c.Request().Context(),
		&cms_proto.ListCategoriesRequest{
			Page:    req.Page,
			PerPage: req.Limit,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) CreatePost(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin && !claims.Author {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	var post Post

	if err := c.Bind(&post); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	nPost := &cms_proto.CreatePostRequest{
		Post: &cms_proto.Post{
			CategoryId:    &cms_proto.UUID{Value: post.CategroyId},
			AuthorId:      &cms_proto.UUID{Value: post.AuthorId},
			Status:        post.Status,
			Title:         post.Title,
			Slug:          post.Slug,
			CoverImage:    post.CoverImage,
			UpdatedDate:   post.UpdatedDate,
			PublishedDate: post.PublishedDate,
			Description:   post.Description,
			Content:       post.Content,
		},
	}

	resp, err := s.CmsClient.CreatePost(c.Request().Context(), nPost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, resp)
}

func (s *CmsServer) GetPost(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id was not provided",
		})
	}

	resp, err := s.CmsClient.GetPost(
		c.Request().Context(),
		&cms_proto.GetPostRequest{
			PostId: &cms_proto.UUID{Value: id},
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *CmsServer) UpdatePost(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin && !claims.Author {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	var post Post

	if err := c.Bind(&post); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	if post.Id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "post id was not provided",
		})
	}

	uPost := &cms_proto.UpdatePostRequest{
		PostId: &cms_proto.UUID{Value: post.Id},
		Post: &cms_proto.Post{
			PostId:        &cms_proto.UUID{Value: post.Id},
			CategoryId:    &cms_proto.UUID{Value: post.CategroyId},
			AuthorId:      &cms_proto.UUID{Value: post.AuthorId},
			Status:        post.Status,
			Title:         post.Title,
			Slug:          post.Slug,
			CoverImage:    post.CoverImage,
			UpdatedDate:   post.UpdatedDate,
			PublishedDate: post.PublishedDate,
			Description:   post.Description,
			Content:       post.Content,
		},
	}
	resp, err := s.CmsClient.UpdatePost(c.Request().Context(), uPost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, resp)
}

func (s *CmsServer) DeletePost(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)
	if !claims.Admin && !claims.Author {
		return c.JSON(http.StatusUnauthorized, ErrResponse{
			Message: "not authorized for this operations",
		})
	}

	id := c.Param("id")

	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "id was not provided",
		})
	}

	resp, err := s.CmsClient.DeletePost(
		c.Request().Context(),
		&cms_proto.DeletePostRequest{PostId: &cms_proto.UUID{Value: id}},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusNoContent, resp)
}

func (s *CmsServer) ListPosts(c echo.Context) error {
	var req PaginatedReq

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid reqest",
		})
	}

	resp, err := s.CmsClient.ListPosts(
		c.Request().Context(),
		&cms_proto.ListPostsRequest{Page: req.Page, PerPage: req.Limit},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

type PaginatedPostCategories struct {
	CategoryID string `json:"category_id"`
	Limit      int32  `json:"limit"`
	Page       int32  `json:"page"`
}

func (s *CmsServer) GetCategoryPosts(c echo.Context) error {
	var req PaginatedPostCategories

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	resp, err := s.CmsClient.GetCategoryPosts(
		c.Request().Context(),
		&cms_proto.GetCategoryPostsRequest{
			CategoryId: &cms_proto.UUID{Value: req.CategoryID},
			Page:       req.Page,
			PerPage:    req.Limit,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

type PaginatedAuthorPosts struct {
	AuthorId string `json:"author_id"`
	Limit    int32  `json:"limit"`
	Page     int32  `json:"page"`
}

func (s *CmsServer) GetAuthorPosts(c echo.Context) error {
	var req PaginatedAuthorPosts
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	resp, err := s.CmsClient.GetAuthorPosts(
		c.Request().Context(),
		&cms_proto.GetAuthorPostsRequest{
			AuthorId: &cms_proto.UUID{Value: req.AuthorId},
			Page:     req.Page,
			PerPage:  req.Limit,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

type PaginatedAuthorCategoryPosts struct {
	CategoryID string `json:"category_id"`
	AuthorId   string `json:"author_id"`
	Limit      int32  `json:"limit"`
	Page       int32  `json:"page"`
}

func (s *CmsServer) GetAuthorCategoryPosts(c echo.Context) error {
	var req PaginatedAuthorCategoryPosts
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	resp, err := s.CmsClient.GetAuthorCategoryPosts(
		c.Request().Context(),
		&cms_proto.GetAuthorCategoryPostsRequest{
			AuthorId:   &cms_proto.UUID{Value: req.AuthorId},
			CategoryId: &cms_proto.UUID{Value: req.CategoryID},
			Page:       req.Page,
			PerPage:    req.Limit,
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}
