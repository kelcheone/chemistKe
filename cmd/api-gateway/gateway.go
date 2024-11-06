package main

import (
	"log"

	authservice "github.com/kelcheone/chemistke/cmd/api-gateway/auth"
	userRoute "github.com/kelcheone/chemistke/cmd/api-gateway/routes/users"
	"github.com/kelcheone/chemistke/cmd/utils"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	userClient user_proto.UserServiceClient
}

func NewServer(userClient user_proto.UserServiceClient) *Server {
	return &Server{
		userClient: userClient,
	}
}

func main() {
	userServer, CloseUserConn, err := userRoute.ConnectServer("localhost:8090")
	if err != nil {
		log.Fatal(err)
	}

	defer CloseUserConn()

	e := echo.New()

	e.Use(middleware.Logger())

	v1 := e.Group("/api/v1")

	users := v1.Group("/users")
	users.POST("", userServer.CreateUser)
	users.GET("/get-user", userServer.GetUser)
	users.GET("", userServer.GetUsers, utils.AuthMiddleware)
	users.GET("/get-user-by-email", userServer.GetUserByEmail)
	users.PATCH("", userServer.UpdateUser, utils.AuthMiddleware)
	users.DELETE("", userServer.DeleteUser, utils.AuthMiddleware)

	auth := v1.Group("/auth")
	auth.POST("/login", func(c echo.Context) error {
		user := authservice.User{
			Client: userServer.UserClient,
		}

		return user.Login(c)
	})

	e.Logger.Fatal(e.Start(":9090"))
}
