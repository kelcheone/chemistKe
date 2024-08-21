package userservice

import (
	"context"

	"github.com/kelcheone/chemistke/internal/database"
	pb "github.com/kelcheone/chemistke/pkg/grpc/user"
)

type UserService struct {
	db database.DB
	pb.UnimplementedUserServiceServer
}

func NewService(db database.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) AddUser(ctx context.Context, req *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	user := req.User
	stmt := `INSERT INTO users (name, email, phone) VALUES ($1, $2, $3) RETURNING id`
	type result struct {
		id string
	}

	res := s.db.QueryRow(stmt, user.Name, user.Email, user.Phone)

	var r result
	if err := res.Scan(&r.id); err != nil {
		return nil, err
	}

	response := &pb.AddUserResponse{
		Message: "User Added sucessfully",
		Id:      &pb.UUID{Value: r.id},
	}
	return response, nil
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	stmt := `SELECT id, name, email, phone FROM users WHERE id=$1`
	row := s.db.QueryRow(stmt, req.Id.Value)

	var gUser pb.User

	var userId string

	err := row.Scan(&userId, &gUser.Name, &gUser.Email, &gUser.Phone)
	if err != nil {
		return nil, err
	}
	gUser.Id = &pb.UUID{Value: userId}

	return &pb.GetUserResponse{User: &gUser}, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserByEmailResponse, error) {
	stmt := `SELECT id, name, email, phone FROM users WHERE email=$1`
	row := s.db.QueryRow(stmt, req.Email)

	var gUser pb.User

	var userId string

	err := row.Scan(&userId, &gUser.Name, &gUser.Email, &gUser.Phone)
	if err != nil {
		return nil, err
	}
	gUser.Id = &pb.UUID{Value: userId}

	return &pb.GetUserByEmailResponse{User: &gUser}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	stmt := `UPDATE users SET name=$1, email=$2, phone=$3 WHERE id=$4`
	tUser := req.User
	_, err := s.db.Exec(stmt, tUser.Name, tUser.Email, tUser.Phone, tUser.Id.Value)
	if err != nil {
		return nil, err
	}
	return &pb.UpdateUserResponse{
		Message: "user updated sucessfully",
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	stmt := `DELETE FROM users WHERE id=$1`
	_, err := s.db.Exec(stmt, req.Id.Value)
	if err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{
		Message: "sucessfully deleted user",
	}, nil
}

func (s *UserService) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	limit := req.Limit
	page := req.Page
	stmt := `SELECT id,name, email, phone FROM users LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(stmt, limit, page)
	if err != nil {
		return nil, err
	}

	var users []*pb.User

	for rows.Next() {
		var user pb.User
		var userId string
		err := rows.Scan(&userId, &user.Name, &user.Email, &user.Phone)
		if err != nil {
			return nil, err
		}
		user.Id = &pb.UUID{Value: userId}
		users = append(users, &user)
	}
	return &pb.GetUsersResponse{
		Users: users,
	}, nil
}
