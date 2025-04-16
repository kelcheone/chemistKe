package authservice

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kelcheone/chemistke/cmd/utils"
	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
	"github.com/labstack/echo/v4"
)

type User struct {
	Id       string `json:"id"`       // User unique identifier
	Name     string `json:"name"`     // User full name
	Email    string `json:"email"`    // User email address
	Phone    string `json:"phone"`    // User phone number
	Password string `json:"password"` // User password (will be compared with stored hash)
	Role     string `json:"role"`     // User role (admin, customer, etc.)
	Client   user_proto.UserServiceClient
}

// LoginRequest represents the expected request body for the login endpoint
type LoginRequest struct {
	Id       string `json:"id"       example:"1bf447b8-a129-42a2-b11e-684a801568ff"` // User ID
	Password string `json:"password" example:"securepassword123"`                    // User password
}

// LoginResponse represents the response from a successful login
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT authentication token
}

// ErrResponse represents an error response
type ErrResponse struct {
	Message string `json:"message" example:"invalid credentials"` // Error message
}

// NewUserClient initializes a new user client with the gRPC service
func NewUserClient(client user_proto.UserServiceClient) {
	// Implementation omitted
}

// Login godoc
// @Summary User login
// @Description Authenticates a user and returns a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User credentials"
// @Success 202 {object} LoginResponse "Successfully authenticated"
// @Failure 400 {object} ErrResponse "Bad request - invalid input"
// @Failure 401 {object} ErrResponse "Unauthorized - invalid credentials"
// @Failure 500 {object} ErrResponse "Internal server error"
// @Router /auth/login [post]
func (u *User) Login(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	_ = ctx

	defer cancel()

	var user User

	if err := c.Bind(&user); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			ErrResponse{Message: "bad request"},
		)
	}

	gUserResp, err := u.Client.GetUser(
		ctx,
		&user_proto.GetUserRequest{Id: &user_proto.UUID{Value: user.Id}},
	)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: fmt.Sprintf("%v", err.Error())},
		)
	}

	hashedPassword := gUserResp.User.Password

	err = utils.ComparePassword(user.Password, hashedPassword)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: fmt.Sprintf("%v", err.Error())},
		)
	}

	tokenString, err := utils.CreateToken(
		user.Id,
		user.Name,
		gUserResp.User.Role.String(),
	)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: "Could not login, Wrong password provided"},
		)
	}

	type response struct {
		Token string `json:"token"`
	}

	return c.JSON(http.StatusAccepted, response{Token: tokenString})
}
