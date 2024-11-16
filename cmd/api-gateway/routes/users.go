package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kelcheone/chemistke/cmd/utils"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type GetUserResponse struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

type ErrResponse struct {
	Message string `json:"error"`
}

type UserServer struct {
	UserClient user_proto.UserServiceClient
}

func ConnectUserServer(link string) (*UserServer, func(), error) {
	userConn, err := grpc.NewClient(
		link,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"Faild to connect to the user service: %v",
			err,
		)
	}

	return &UserServer{
			UserClient: user_proto.NewUserServiceClient(userConn),
		}, func() {
			userConn.Close()
		}, nil
}

func (s *UserServer) CreateUser(c echo.Context) error {
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()
	type response struct {
		Message string `json:"message"`
	}
	err := c.Bind(&user)
	if err != nil {
		return c.JSON(http.StatusBadRequest, response{
			Message: "bad request",
		})
	}

	pbUSer := &user_proto.User{
		Name:     user.Name,
		Email:    user.Email,
		Phone:    user.Phone,
		Password: user.Password,
	}
	if user.Role == "ADMIN" {
		pbUSer.Role = user_proto.UserRoles_ADMIN
	} else {
		pbUSer.Role = user_proto.UserRoles_USER
	}

	fmt.Printf("%+v\n", pbUSer)
	res, err := s.UserClient.AddUser(
		ctx,
		&user_proto.AddUserRequest{User: pbUSer},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: fmt.Sprintf("%+v", err.Error()),
		})
	}
	return c.JSON(http.StatusCreated, res)
}

func (s *UserServer) GetUser(c echo.Context) error {
	var user User

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	userReq := user_proto.GetUserRequest{
		Id: &user_proto.UUID{
			Value: user.Id,
		},
	}

	gUser, err := s.UserClient.GetUser(context.TODO(), &userReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	response := GetUserResponse{
		Id:    gUser.User.Id.Value,
		Name:  gUser.User.Name,
		Email: gUser.User.Email,
		Phone: gUser.User.Phone,
		Role:  gUser.User.Role.String(),
	}
	return c.JSON(http.StatusOK, response)
}

func (s *UserServer) GetUserByEmail(c echo.Context) error {
	var user User

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: err.Error(),
		})
	}

	userReq := user_proto.GetUserByEmailRequest{
		Email: user.Email,
	}

	gUser, err := s.UserClient.GetUserByEmail(context.TODO(), &userReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}

	response := GetUserResponse{
		Id:    gUser.User.Id.Value,
		Name:  gUser.User.Name,
		Email: gUser.User.Email,
		Phone: gUser.User.Phone,
		Role:  gUser.User.Role.String(),
	}
	return c.JSON(http.StatusOK, response)
}

func (s *UserServer) UpdateUser(c echo.Context) error {
	var user User

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "Invalid input",
		})
	}

	userClaims := utils.ExtractClaimsFromRequest(c)
	if userClaims.Id != user.Id {
		return c.JSON(
			http.StatusUnauthorized,
			ErrResponse{Message: "can't update this record"},
		)
	}

	var role user_proto.UserRoles

	if user.Role == "ADMIN" {
		role = user_proto.UserRoles_ADMIN
	} else {
		role = user_proto.UserRoles_USER
	}

	req := &user_proto.UpdateUserRequest{
		User: &user_proto.User{
			Id:    &user_proto.UUID{Value: user.Id},
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
			Role:  role,
		},
	}

	resp, err := s.UserClient.UpdateUser(context.TODO(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: strings.TrimSpace(
				err.Error(),
			), // strings.Split(strings.Split(err.Error(), ",")[1], ":")[1],

		})
	}
	return c.JSON(http.StatusNoContent, resp)
}

func (s *UserServer) DeleteUser(c echo.Context) error {
	var user User

	if err := c.Bind(&user); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			ErrResponse{Message: "invalid request"},
		)
	}

	userClaims := utils.ExtractClaimsFromRequest(c)
	if userClaims.Id != user.Id {
		return c.JSON(
			http.StatusUnauthorized,
			ErrResponse{Message: "can't update this record"},
		)
	}
	resp, err := s.UserClient.DeleteUser(
		context.TODO(),
		&user_proto.DeleteUserRequest{Id: &user_proto.UUID{Value: user.Id}},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusAccepted, resp)
}

func (s *UserServer) GetUsers(c echo.Context) error {
	type reqType struct {
		Page  int
		Limit int
	}

	var req reqType

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	resp, err := s.UserClient.GetUsers(
		context.TODO(),
		&user_proto.GetUsersRequest{
			Page:  int32(req.Page),
			Limit: int32(req.Limit),
		},
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	var responses []GetUserResponse

	for _, gUser := range resp.Users {
		response := GetUserResponse{
			Id:    gUser.Id.Value,
			Name:  gUser.Name,
			Email: gUser.Email,
			Phone: gUser.Phone,
			Role:  gUser.Role.String(),
		}
		responses = append(responses, response)
	}

	type Response struct {
		Users []GetUserResponse `json:"users"`
	}
	fResp := Response{
		Users: responses,
	}

	return c.JSON(http.StatusOK, fResp)
}
