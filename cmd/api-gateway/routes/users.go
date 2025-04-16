package routes

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kelcheone/chemistke/cmd/utils"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// User represents the data needed to create a suser
type User struct {
	Id       string `json:"id"       example:"42cef6ad-1b39-4708-aa3f-a0c485f70db3"`
	Name     string `json:"name"     example:"Jane Doe"                             binding:"required"`
	Email    string `json:"email"    example:"jane.doe@example.com"                 binding:"required"`
	Phone    string `json:"phone"    example:"+254722000000"                        binding:"required"`
	Password string `json:"password" example:"12345"                                binding:"required"`
	Role     string `json:"role"     example:"USER"`
}

// GetUserResponse represents the user data returned by the endpoint
type GetUserResponse struct {
	Id    string `json:"id"    example:"cef6ad-1b39-4708-aa3f-a0c485f70db3"`
	Name  string `json:"name"  example:"Jane Doe"`
	Email string `json:"email" example:"jane.doe@example.com"`
	Phone string `json:"phone" example:"+254722000000"`
	Role  string `json:"role"  example:"USER"`
}

// HTTPError represents an error response
type HTTPError struct {
	Message string `json:"error"`
}

// ErrResponse represents an error response
type ErrResponse struct {
	Message string `json:"error"`
}

// UserServer handles user-related API endpoints
type UserServer struct {
	UserClient user_proto.UserServiceClient
}

// ConnectUserServer creates a connection to the user gRPC service
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

// CreateUser godoc
// @Summary Create a new user
// @Description Register a new user to the system
// @Tags Users
// @Accept json
// @Produce json
// @Param user body User true "User information to create"
// @Success 201 {object} GetUserResponse "Successfully created user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /users [post]
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

// GetUser godoc
// @Summary Get a user by ID
// @Description Get user details by user ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Success 200 {object} GetUserResponse "Successfully retrieved user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 404 {object} HTTPError "User not found"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /users/get-user [get]
func (s *UserServer) GetUser(c echo.Context) error {
	// Get the ID from query parameters instead of binding JSON
	id := c.QueryParam("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "missing user ID",
		})
	}

	userReq := user_proto.GetUserRequest{
		Id: &user_proto.UUID{
			Value: id,
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

// GetUserByEmail godoc
// @Summary Get a user by email
// @Description Get user details by email address
// @Tags Users
// @Accept json
// @Produce json
// @Param email query string true "User Email"
// @Success 200 {object} GetUserResponse "Successfully retrieved user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 404 {object} HTTPError "User not found"
// @Failure 500 {object} HTTPError "Internal server error"
// @Router /users/get-user-by-email [get]
func (s *UserServer) GetUserByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "bad email request",
		})
	}
	userReq := user_proto.GetUserByEmailRequest{
		Email: email,
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

// UpdateUser godoc
// @Summary Update user details
// @Description Update existing user information
// @Tags Users
// @Accept json
// @Produce json
// @Param user body User true "Updated user information"
// @Security ApiKeyAuth
// @Success 200 {object} GetUserResponse "Successfully updated user"
// @Failure 400 {object} HTTPError "Invalid input data"
// @Failure 401 {object} HTTPError "Unauthorized"
// @Failure 404 {object} HTTPError "User not found"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /users [patch]
func (s *UserServer) UpdateUser(c echo.Context) error {
	var user User

	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "Invalid input",
		})
	}

	userClaims := utils.ExtractClaimsFromRequest(c)
	log.Printf("UserId %v vs Claims UserId %v", user.Id, userClaims.Id)
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

// DeleteUserResponse represents the response from the delete user operation
type DeleteUserResponse struct {
	Success bool   `json:"success"           example:"true"`
	Message string `json:"message,omitempty" example:"User deleted successfully"`
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete an existing user
// @Tags Users
// @Accept json
// @Produce json
// @Param user body User true "User ID to delete"
// @Security ApiKeyAuth
// @Success 202 {object} DeleteUserResponse "Successfully deleted user"
// @Failure 400 {object} ErrResponse "Invalid input data"
// @Failure 401 {object} ErrResponse "Unauthorized"
// @Failure 500 {object} ErrResponse "Internal server error"
// @Security BearerAuth
// @Router /users [delete]
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

// GetUsersRequest represents the pagination request
type GetUsersRequest struct {
	Page  int `json:"page"  example:"1"`
	Limit int `json:"limit" example:"50"`
}

// GetUsers godoc
// @Summary Get all users
// @Description Get a list of all users
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int true "GetUsersRequest Page"
// @Param limit query int tru "GetUsersRequest Limit"
// @Success 200 {array} GetUserResponse "Successfully retrieved users"
// @Failure 401 {object} HTTPError "Unauthorized"
// @Failure 500 {object} HTTPError "Internal server error"
// @Security BearerAuth
// @Router /users [get]
func (s *UserServer) GetUsers(c echo.Context) error {
	var req GetUsersRequest

	page := c.QueryParam("page")

	n_page, err := strconv.Atoi(page)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}
	req.Page = n_page

	limit := c.QueryParam("limit")

	n_limit, err := strconv.Atoi(limit)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse{
			Message: "invalid request",
		})
	}

	req.Limit = n_limit

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
