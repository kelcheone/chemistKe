package cmsClient

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	cms_proto "github.com/kelcheone/chemistke/pkg/grpc/cms"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Init() error {
	conn, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// create category -> user -> author -> post
	cUser := user_proto.NewUserServiceClient(conn)
	c := cms_proto.NewCmsServiceClient(conn)

	// create category
	newCategory := &cms_proto.Category{
		Name:        gofakeit.ProductCategory(),
		Slug:        gofakeit.URL(),
		Description: gofakeit.LoremIpsumSentence(50),
	}

	createCategoryRes, err := c.CreateCategory(ctx, &cms_proto.CreateCategoryRequest{Category: newCategory})
	if err != nil {
		return err
	}

	fmt.Printf("Category Created: %v\n", createCategoryRes.CategoryId)

	// create user

	newUser := user_proto.User{
		Name:     gofakeit.Name(),
		Email:    gofakeit.Email(),
		Role:     user_proto.UserRoles_AUTHOR,
		Phone:    gofakeit.Phone(),
		Password: gofakeit.Password(true, true, true, true, true, 12),
	}

	user, err := cUser.AddUser(ctx, &user_proto.AddUserRequest{User: &newUser})
	if err != nil {
		return err
	}

	fmt.Printf("User created: %v %s\n", user.Id, user.Message)

	// create an author
	newAuthor := &cms_proto.Author{
		Bio:    gofakeit.LoremIpsumSentence(60),
		Url:    gofakeit.URL(),
		Avatar: gofakeit.URL(),
		UserId: &cms_proto.UUID{Value: user.Id.Value},
	}

	author, err := c.CreateAuthor(ctx, &cms_proto.CreateAuthorRequest{Author: newAuthor})
	if err != nil {
		return err
	}
	fmt.Printf("Author created: %v\n", author.AuthorId)

	// create post
	newPost := &cms_proto.Post{
		PublishedDate: gofakeit.Date().Format("2006-01-02 15:04"),
		UpdatedDate:   gofakeit.Date().Format("2006-01-02 15:04"),
		CoverImage:    gofakeit.URL(),
		Title:         gofakeit.BookTitle(),
		Description:   gofakeit.LoremIpsumSentence(60),
		Slug:          gofakeit.URL(),
		Content:       gofakeit.LoremIpsumParagraph(6, 6, 100, " "),
		Status:        "draft",
		AuthorId:      author.AuthorId,
		CategoryId:    createCategoryRes.CategoryId,
	}

	postId, err := c.CreatePost(ctx, &cms_proto.CreatePostRequest{Post: newPost})
	if err != nil {
		return err
	}

	fmt.Printf("Created Post: %v\n", postId.PostId)

	// Update Post
	newPost.PostId = postId.PostId
	newPost.Slug = gofakeit.URL()
	newPost.Content = gofakeit.LoremIpsumParagraph(6, 6, 300, "    ")
	newPost.Status = "pupblished"

	updatePostRes, err := c.UpdatePost(ctx, &cms_proto.UpdatePostRequest{PostId: newPost.PostId, Post: newPost})
	if err != nil {
		return err
	}

	fmt.Printf("Updated user: %v\n", updatePostRes.PostId)

	// GET post
	getPostRes, err := c.GetPost(ctx, &cms_proto.GetPostRequest{PostId: newPost.PostId})
	if err != nil {
		return err
	}
	fmt.Printf("Got Post: %s %s %s", getPostRes.Post.Slug, getPostRes.Post.Title, getPostRes.Post.Description)

	// Get Posts
	listRes, err := c.ListPosts(ctx, &cms_proto.ListPostsRequest{Page: 1, PerPage: 30})
	if err != nil {
		return err
	}

	fmt.Println("------------------------POSTS---------------------")
	for i, gPost := range listRes.Posts {
		fmt.Printf("%d ----------> %v\n", i, gPost.Title)
	}

	// get categories
	listCatRes, err := c.ListCategories(ctx, &cms_proto.ListCategoriesRequest{Page: 1, PerPage: 30})
	if err != nil {
		return err
	}
	fmt.Println("------------------------CATEGORIES---------------------")
	for i, gCategory := range listCatRes.Categories {
		fmt.Printf("%d ----------> %v\n", i, gCategory.Name)
	}

	// get authors
	listAuthors, err := c.ListAuthors(ctx, &cms_proto.ListAuthorsRequest{Page: 1, PerPage: 30})
	if err != nil {
		return err
	}

	fmt.Println("------------------------AUTHORS---------------------")
	for i, gAuthor := range listAuthors.Authors {
		fmt.Printf("%d ----------> %v\n", i, gAuthor.UserId)
	}
	return nil
}
