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
	Id       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Client   user_proto.UserServiceClient
}

type ErrResponse struct {
	Message string `json:"message"`
}

var secretKey = []byte("secret-key")

func NewUserClient(client user_proto.UserServiceClient) {
}

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

	tokenString, err := utils.CreateToken(user.Id, user.Name, false)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			ErrResponse{Message: fmt.Sprintf("%v", err.Error())},
		)
	}

	type response struct {
		Token string `json:"token"`
	}

	return c.JSON(http.StatusAccepted, response{Token: tokenString})
}
