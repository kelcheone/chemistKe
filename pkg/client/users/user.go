package userClient

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	user_proto "github.com/kelcheone/chemistke/pkg/grpc/user"
)

func Init() error {
	conn, err := grpc.NewClient(
		"localhost:8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	// fmt.Printf("%+v\n", conn)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := user_proto.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	//-------------ADD USER ----------

	user := &user_proto.AddUserRequest{
		User: &user_proto.User{
			Name:     gofakeit.Name(),
			Email:    gofakeit.Email(),
			Phone:    gofakeit.Phone(),
			Password: gofakeit.Password(true, true, true, true, false, 12),
			Role:     user_proto.UserRoles_ADMIN,
		},
	}

	res, err := c.AddUser(ctx, user)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", res)

	// -------------- GET-USER------------
	getUserReq := &user_proto.GetUserRequest{
		Id: res.Id,
	}

	getUserRes, err := c.GetUser(ctx, getUserReq)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", getUserRes)

	// --------------- UPDATE USER -----------

	nUser := &user_proto.UpdateUserRequest{
		User: &user_proto.User{
			Id:       res.Id,
			Name:     gofakeit.Name(),
			Email:    gofakeit.Email(),
			Phone:    gofakeit.Phone(),
			Password: gofakeit.Password(true, true, true, true, false, 12),
		},
	}

	updateRes, err := c.UpdateUser(ctx, nUser)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", updateRes)

	//----------------- DELETE USER ---------------
	delReq := &user_proto.DeleteUserRequest{Id: res.Id}
	delRes, err := c.DeleteUser(ctx, delReq)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", delRes)

	//---------------- GET USERS ----------------
	getUsersReq := &user_proto.GetUsersRequest{
		Limit: 5,
		Page:  1,
	}

	getUsersResp, err := c.GetUsers(ctx, getUsersReq)
	if err != nil {
		return err
	}
	for i, user := range getUsersResp.Users {
		fmt.Printf("%d-> %+v\n", i, user)
	}

	for range 20 {
		user := &user_proto.AddUserRequest{
			User: &user_proto.User{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Phone:    gofakeit.Phone(),
				Password: gofakeit.Password(true, true, true, true, false, 12),
				Role:     user_proto.UserRoles_ADMIN,
			},
		}

		_, err := c.AddUser(context.TODO(), user)
		if err != nil {
			log.Fatal(err)
		}

	}

	return nil
}

func CreateUsers(
	ctx context.Context,
	c user_proto.UserServiceClient,
) {
	for range 20 {
		user := &user_proto.AddUserRequest{
			User: &user_proto.User{
				Name:     gofakeit.Name(),
				Email:    gofakeit.Email(),
				Phone:    gofakeit.Phone(),
				Password: gofakeit.Password(true, true, true, true, false, 12),
				Role:     user_proto.UserRoles_ADMIN,
			},
		}

		_, err := c.AddUser(ctx, user)
		if err != nil {
			log.Fatal(err)
		}

	}
}

// ------------------Add 20 Users
