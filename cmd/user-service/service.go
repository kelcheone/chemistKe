package userservice

import (
	"context"
	"database/sql"

	"github.com/kelcheone/chemistke/cmd/utils"
	"github.com/kelcheone/chemistke/internal/database"
	"github.com/kelcheone/chemistke/pkg/codes"
	pb "github.com/kelcheone/chemistke/pkg/grpc/user"
	"github.com/kelcheone/chemistke/pkg/status"
)

type UserService struct {
	db database.DB
	pb.UnimplementedUserServiceServer
}

func NewService(db database.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) AddUser(
	ctx context.Context,
	req *pb.AddUserRequest,
) (*pb.AddUserResponse, error) {
	user := req.User
	stmt := `INSERT INTO users (name, email, phone, role, password) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	type result struct {
		id string
	}

	hashedPassword, err := utils.Hash(user.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res := s.db.QueryRow(
		stmt,
		user.Name,
		user.Email,
		user.Phone,
		user.Role,
		hashedPassword,
	)

	var r result
	if err := res.Scan(&r.id); err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "could not create user")
	}

	response := &pb.AddUserResponse{
		Message: "User Added sucessfully",
		Id:      &pb.UUID{Value: r.id},
	}
	return response, nil
}

func (s *UserService) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.GetUserResponse, error) {
	stmt := `SELECT id, name, email, phone, role, password FROM users WHERE id=$1`
	row := s.db.QueryRow(stmt, req.Id.Value)

	var gUser pb.User

	var userId string

	err := row.Scan(
		&userId,
		&gUser.Name,
		&gUser.Email,
		&gUser.Phone,
		&gUser.Role,
		&gUser.Password,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "error fetching user")
	}
	gUser.Id = &pb.UUID{Value: userId}

	return &pb.GetUserResponse{User: &gUser}, nil
}

func (s *UserService) GetUserByEmail(
	ctx context.Context,
	req *pb.GetUserByEmailRequest,
) (*pb.GetUserByEmailResponse, error) {
	stmt := `SELECT id, name, email, phone, role FROM users WHERE email=$1`
	row := s.db.QueryRow(stmt, req.Email)

	var gUser pb.User

	var userId string

	err := row.Scan(
		&userId,
		&gUser.Name,
		&gUser.Email,
		&gUser.Phone,
		&gUser.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(
				codes.NotFound,
				"use with email %s not found",
				req.Email,
			)
		}
		return nil, status.Errorf(
			codes.Internal,
			"error fetching user with email %s",
			req.Email,
		)
	}
	gUser.Id = &pb.UUID{Value: userId}

	return &pb.GetUserByEmailResponse{User: &gUser}, nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	req *pb.UpdateUserRequest,
) (*pb.UpdateUserResponse, error) {
	stmt := `UPDATE users SET name=$1, email=$2, phone=$3, role=$4 WHERE id=$5`
	tUser := req.User

	if tUser.Id.Value == "" {
		return nil, status.Errorf(codes.Aborted, "Id was not provided")
	}
	_, err := s.db.Exec(
		stmt,
		tUser.Name,
		tUser.Email,
		tUser.Phone,
		tUser.Role,
		tUser.Id.Value,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "could not update user")
	}
	return &pb.UpdateUserResponse{
		Message: "user updated sucessfully",
	}, nil
}

func (s *UserService) DeleteUser(
	ctx context.Context,
	req *pb.DeleteUserRequest,
) (*pb.DeleteUserResponse, error) {
	stmt := `DELETE FROM users WHERE id=$1`
	_, err := s.db.Exec(stmt, req.Id.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "could not delete user")
	}
	return &pb.DeleteUserResponse{
		Message: "sucessfully deleted user",
	}, nil
}

func (s *UserService) GetUsers(
	ctx context.Context,
	req *pb.GetUsersRequest,
) (*pb.GetUsersResponse, error) {
	limit := req.Limit
	page := req.Page
	stmt := `SELECT id,name, email, phone, role FROM users LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(stmt, limit, page)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not fetch users")
	}

	var users []*pb.User

	for rows.Next() {
		var user pb.User
		var userId string
		err := rows.Scan(
			&userId,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.Role,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error scanning user")
		}
		user.Id = &pb.UUID{Value: userId}
		users = append(users, &user)
	}
	return &pb.GetUsersResponse{
		Users: users,
	}, nil
}
