package routes

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/kelcheone/chemistke/cmd/utils"
	cms_proto "github.com/kelcheone/chemistke/pkg/grpc/cms"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Post represents the required post format and what is also returned
type Post struct {
	Id            string `json:"id"             example:"1bf447b8-a129-42a2-b11e-684a801568ff"`
	PublishedDate string `json:"published_date" example:"2024-11-16 10:30:00"`
	UpdatedDate   string `json:"updated_date"   example:"2024-11-16 10:30:00"`
	CoverImage    string `json:"cover_image"    example:"https://example.com/images/crypto-future.jpg"`
	Title         string `json:"title"          example:"Why ozempic is good for you."                 binding:"required"`
	Description   string `json:"description"    example:"10 benefits of ozempic"                       binding:"required"`
	Slug          string `json:"slug"           example:"why-ozempic-is-good-for-you"                  binding:"required"`
	Content       string `json:"content"        example:"lorem ipsum"                                  binding:"required"`
	Status        string `json:"status"         example:"draft"                                        binding:"required"`
	AuthorId      string `json:"author_id"      example:"42cef6ad-1b39-4708-aa3f-a0c485f70db3"         binding:"required"`
	CategroyId    string `json:"category_id"    example:"42cef6ad-1b39-4708-aa3f-a0c485f70db3"         binding:"required"`
}

// Author represents a given author
type Author struct {
	Id     string `json:"id"      example:"42cef6ad-1b39-4708-aa3f-a0c485f70db3"`
	Bio    string `json:"bio"     example:"I like writing about meds"             binding:"required"`
	Avatar string `json:"avatar"  example:"https://example.com/images/avatar.png"`
	Url    string `json:"url"     example:"https://mywebsite.com"`
	UserId string `json:"user_id" example:"42cef6ad-1b39-4708-aa3f-a0c485f70db3"  binding:"required"`
}

// Category represents category to which the content refers to
type Category struct {
	Id          string `json:"id"          example:"42cef6ad-1b39-4708-aa3f-a0c485f70db3"`
	Name        string `json:"name"        example:"Antibiotics"                            binding:"required"`
	Slug        string `json:"slug"        example:"antibiotics"                            binding:"required"`
	Description string `json:"description" example:"All you need to know about antibiotics" binding:"required"`
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

// CreateAuthor godoc
// @Summary Registers a new Author.
// @Description Creates a new Author in the system.
// @Tags Content
// @Accept json
// @Produce json
// @Param author body Author true "Author to create"
// @Success 201 {object} Author "Author created Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Security BearerAuth
// @Router /cms/authors [post]
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

// GetAuthor godoc
// @Summary gets an Author.
// @Description Gets an author based on a provided id.
// @Tags Content
// @Accept json
// @Produce json
// @Param id path string true "Author ID"
// @Success 200 {object} Author "Fetched Author Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Router /cms/authors/{id} [get]
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

// UpdateAuthor godoc
// @Summary Updates an Author.
// @Description Updates an Author in the system.
// @Tags Content
// @Accept json
// @Produce json
// @Param author body Author true "Author to update"
// @Success 204 {object} Author "Author updated Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Security BearerAuth
// @Router /cms/authors [patch]
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

// DeleteAuthor godoc
// @Summary deletes an Author.
// @Description deletes an author based on a provided id.
// @Tags Content
// @Accept json
// @Produce json
// @Param id path string true "Author ID"
// @Success 204 {object} Author "Fetched Author Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Security BearerAuth
// @Router /cms/authors/{id} [delete]
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

// ListAuthors godoc
// @Summary lists all Authors.
// @Description lists all authors in the system
// @Tags Content
// @Accept json
// @Produce json
// @Param page query int true "Page number"
// @Param limit query int true "Limit"
// @Success 200 {object} Author "Fetched Author Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Router /cms/authors [get]
func (s *CmsServer) ListAuthors(c echo.Context) error {
	page := c.QueryParam("page")
	log.Printf("The current page is %v\n", page)

	int_page, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	limit := c.QueryParam("limit")
	log.Printf("The required limit is: %v\n", limit)

	int_limit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}
	resp, err := s.CmsClient.ListAuthors(
		c.Request().Context(),
		&cms_proto.ListAuthorsRequest{
			Page:    int32(int_page),
			PerPage: int32(int_limit),
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateCategory godoc
// @Summary Registers a new cateogory.
// @Description Creates a new Category in the system.
// @Tags Content
// @Accept json
// @Produce json
// @Param category body Category true "Category to create"
// @Success 201 {object} Category "Category created Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Security BearerAuth
// @Router /cms/categories [post]
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

// GetCategory godoc
// @Summary Gets a cateogory.
// @Description Gets a Category from the system.
// @Tags Content
// @Accept json
// @Produce json
// @Param id path string true "Category Id"
// @Success 200 {object} Category "Category fetched Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Router /cms/categories/{id} [get]
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

// UpdateCategory godoc
// @Summary Updates a cateogory.
// @Description Updates a Category in the system.
// @Tags Content
// @Accept json
// @Produce json
// @Param category body Category true "Category to create"
// @Success 204 {object} Category "Category updated Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Security BearerAuth
// @Router /cms/categories [patch]
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

// DeleteCategory godoc
// @Summary deletes a cateogory.
// @Description Deletes a Category from the system.
// @Tags Content
// @Accept json
// @Produce json
// @Param id path string true "Category Id"
// @Success 204 {object} Category "Category deleted Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Security BearerAuth
// @Router /cms/categories/{id} [delete]
func (s *CmsServer) DeleteCategory(c echo.Context) error {
	claims := utils.ExtractClaimsFromRequest(c)

	log.Printf("Is admin: %v is Author: %v\n", claims.Admin, claims.Author)
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

// ListCategories godoc
// @Summary lists all categories.
// @Description lists all categories in the system
// @Tags Content
// @Accept json
// @Produce json
// @Param page query int true "PaginatedReq Page"
// @Param limit query int true "PaginatedReq Limit"
// @Success 200 {object} Author "Fetched categories Sucessfully"
// @Failure 400 {object} HTTPError "invalid input data"
// @Failure 500 {object} HTTPError "internal server error"
// @Router /cms/categories [get]
func (s *CmsServer) ListCategories(c echo.Context) error {
	page := c.QueryParam("page")

	int_page, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	limit := c.QueryParam("limit")

	int_limit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}
	resp, err := s.CmsClient.ListCategories(
		c.Request().Context(),
		&cms_proto.ListCategoriesRequest{
			Page:    int32(int_page),
			PerPage: int32(int_limit),
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
