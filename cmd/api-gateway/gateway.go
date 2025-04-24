package main

import (
	"log"

	"github.com/go-playground/validator"
	authservice "github.com/kelcheone/chemistke/cmd/api-gateway/auth"
	routes "github.com/kelcheone/chemistke/cmd/api-gateway/routes"
	"github.com/kelcheone/chemistke/cmd/utils"
	_ "github.com/kelcheone/chemistke/docs"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Server struct {
	userClient user_proto.UserServiceClient
}

func NewServer(userClient user_proto.UserServiceClient) *Server {
	return &Server{
		userClient: userClient,
	}
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

/*
@host chemistke-production.up.railway.app
*/

// @title ChemistKe API
// @version 1.0
// @description API endpoint documentation for the ChemistKe api project.
// @termsOfService http://swagger.io/terms/

// @contact.name ChemistKe Support
// @contact.url https://chemist.co.ke/support
// @contact.email Support@chemist.co.ke

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9090
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Authorization: Bearer {token}"

func main() {
	userServer, CloseUserConn, err := routes.ConnectUserServer("localhost:8090")
	if err != nil {
		log.Fatal(err)
	}

	defer CloseUserConn()

	productsServer, CloseProductConn, err := routes.ConnectProductServer(
		"localhost:8090",
	)
	if err != nil {
		log.Fatal(err)
	}

	defer CloseProductConn()

	ordersServer, CloseOrderConn, err := routes.ConnectOrdersServer(
		"localhost:8090",
	)
	if err != nil {
		log.Fatal(err)
	}

	defer CloseOrderConn()

	cmsServer, CloseCmsConn, err := routes.ConnectCmsServer("localhost:8090")
	if err != nil {
		log.Fatal(err)
	}

	defer CloseCmsConn()

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Requested-With",
		},
		AllowMethods: []string{
			echo.GET,
			echo.HEAD,
			echo.PUT,
			echo.PATCH,
			echo.POST,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
	}))
	e.Use(middleware.Logger())

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := e.Group("/api/v1")

	users := v1.Group("/users")
	users.POST("", userServer.CreateUser)
	users.GET("/get-user", userServer.GetUser)
	users.GET("", userServer.GetUsers, utils.AuthMiddleware())
	users.GET("/get-user-by-email", userServer.GetUserByEmail)
	users.PATCH("", userServer.UpdateUser, utils.AuthMiddleware())
	users.DELETE("", userServer.DeleteUser, utils.AuthMiddleware())

	auth := v1.Group("/auth")
	auth.POST("/login", func(c echo.Context) error {
		user := authservice.User{
			Client: userServer.UserClient,
		}

		return user.Login(c)
	})

	auth.GET("/me", func(c echo.Context) error {
		user := authservice.User{
			Client: userServer.UserClient,
		}

		return user.Me(c)
	}, utils.AuthMiddleware())

	auth.POST("/logout", func(c echo.Context) error {
		user := authservice.User{
			Client: userServer.UserClient,
		}

		return user.Logout(c)
	}, utils.AuthMiddleware())

	products := v1.Group("/products")
	products.POST("", productsServer.CreateProduct, utils.AuthMiddleware())
	products.GET("/:id", productsServer.GetProduct)
	products.GET("", productsServer.GetProducts)
	products.GET("/featured", productsServer.GetFeaturedProducts)
	products.GET("/by-brand/:id", productsServer.GetProductsByBrand)
	products.GET("/by-category/:id", productsServer.GetProductsByCategory)
	products.GET("/by-category/slug/:slug", productsServer.GetProductsByCategorySlug)
	products.GET("/by-subcategory/:id", productsServer.GetProductsBySCategory)
	products.GET("/slug/:slug", productsServer.GetProductBySlug)

	products.POST("/reviews", productsServer.CreateReview, utils.AuthMiddleware())
	products.GET("/ratings/:id", productsServer.GetProductRating)
	products.GET("/reviews/:id", productsServer.GetReview)
	products.GET("/:id/reviews", productsServer.GetReviews)

	// product-category
	products.POST("/categories", productsServer.CreateCategory, utils.AuthMiddleware())
	products.GET("/categories/:id", productsServer.GetCategory)
	products.GET("/categories", productsServer.GetCategories)
	products.GET("/categories/featured", productsServer.GetFeaturedCategories)
	products.PATCH("/categories", productsServer.UpdateCategory, utils.AuthMiddleware())
	products.DELETE("/categories/:id", productsServer.DeleteCategory, utils.AuthMiddleware())
	products.GET("/categories/:id/subcategories", productsServer.GetSubCategories)
	products.GET("/categories/slug/:slug", productsServer.GetCategoryBySlug)
	// product-sub-category
	products.POST("/subcategories", productsServer.CreateSubCategory, utils.AuthMiddleware())
	products.GET("/subcategories/:id", productsServer.GetSubCategory)
	products.GET("/subcategories", productsServer.GetSubCategories)
	products.PATCH("/subcategories", productsServer.UpdateSubCategory, utils.AuthMiddleware())
	products.DELETE("/subcategories/:id", productsServer.DeleteSubCategory, utils.AuthMiddleware())
	products.GET("/subcategories/slug/:slug", productsServer.GetSubCategoryBySlug)
	// product-brand
	products.POST("/brands", productsServer.CreateBrand, utils.AuthMiddleware())
	products.GET("/brands", productsServer.GetBrands)
	products.GET("/brands/:id", productsServer.GetBrand)
	products.PATCH("/brands", productsServer.UpdateBrand, utils.AuthMiddleware())
	products.DELETE("/brands/:id", productsServer.DeleteBrand, utils.AuthMiddleware())

	products.PATCH("", productsServer.UpdateProduct, utils.AuthMiddleware())
	products.DELETE("/:id", productsServer.DeleteProduct, utils.AuthMiddleware())
	products.POST("/images/upload", productsServer.UploadImage)
	products.GET("/images/:id", productsServer.GetProductImages)

	orders := v1.Group("/orders", utils.AuthMiddleware())
	orders.POST("", ordersServer.CreateOrder)
	orders.GET("/:id", ordersServer.GetOrder)
	orders.DELETE("/:id", ordersServer.DeleteOrder)
	orders.GET("/user", ordersServer.GetUserOders)
	orders.GET("", ordersServer.GetOders)
	orders.PATCH("", ordersServer.UpdateOrder)

	cms := v1.Group("/cms")

	authors := cms.Group("/authors")
	authors.POST("", cmsServer.CreateAuthor, utils.AuthMiddleware())
	authors.GET("/:id", cmsServer.GetAuthor)
	authors.PATCH("", cmsServer.UpdateAuthor, utils.AuthMiddleware())
	authors.DELETE("/:id", cmsServer.DeleteAuthor, utils.AuthMiddleware())
	authors.GET("", cmsServer.ListAuthors)

	categories := cms.Group("/categories")
	categories.POST("", cmsServer.CreateCategory, utils.AuthMiddleware())
	categories.GET("/:id", cmsServer.GetCategory)
	categories.GET("", cmsServer.ListCategories)
	categories.PATCH("", cmsServer.UpdateCategory, utils.AuthMiddleware())
	categories.DELETE("/:id", cmsServer.DeleteCategory, utils.AuthMiddleware())

	posts := cms.Group("/posts")
	posts.POST("", cmsServer.CreatePost, utils.AuthMiddleware())
	posts.GET("/:id", cmsServer.GetPost)
	posts.GET("", cmsServer.ListPosts)
	posts.PATCH("", cmsServer.UpdatePost, utils.AuthMiddleware())
	posts.DELETE("/:id", cmsServer.DeletePost, utils.AuthMiddleware())
	posts.GET("/category", cmsServer.GetCategoryPosts)
	posts.GET("/author", cmsServer.GetAuthorPosts)
	posts.GET("/get-by-author-category", cmsServer.GetAuthorCategoryPosts)

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	e.Logger.Fatal(e.Start(":9090"))
}
