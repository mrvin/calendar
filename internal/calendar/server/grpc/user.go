package grpcserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/grpcapi"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateUser(ctx context.Context, userpb *grpcapi.CreateUserRequest) (*emptypb.Empty, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userpb.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		err := fmt.Errorf("create user: generate hash password: %w", err)
		slog.Error(err.Error())
		return nil, err
	}
	user := storage.User{
		Name:         userpb.GetName(),
		HashPassword: string(hashPassword),
		Email:        userpb.GetEmail(),
		Role:         "user",
	}
	if err := s.authService.CreateUser(ctx, &user); err != nil {
		err = fmt.Errorf("create user: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetUser(ctx context.Context, _ *grpcapi.GetUserRequest) (*grpcapi.UserResponse, error) {
	userName := GetUserName(ctx)
	if userName == "" {
		panic("get user: user name is empty")
	}
	user, err := s.authService.GetUser(ctx, userName)
	if err != nil {
		err := fmt.Errorf("get user: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &grpcapi.UserResponse{
		Name:         user.Name,
		HashPassword: user.HashPassword,
		Email:        user.Email,
	}, nil
}

func (s *Server) ListUsers(ctx context.Context, _ *grpcapi.ListUsersRequest) (*grpcapi.ListUsersResponse, error) {
	users, err := s.authService.ListUsers(ctx)
	if err != nil {
		err := fmt.Errorf("get all users: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	pbUsers := make([]*grpcapi.UserResponse, len(users))
	for i, user := range users {
		pbUsers[i] = &grpcapi.UserResponse{
			Name:  user.Name,
			Email: user.Email,
		}
	}

	return &grpcapi.ListUsersResponse{Users: pbUsers}, nil
}

func (s *Server) UpdateUser(ctx context.Context, userpb *grpcapi.UpdateUserRequest) (*emptypb.Empty, error) {
	userName := GetUserName(ctx)
	if userName == "" {
		panic("update user: user name is empty")
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userpb.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		err := fmt.Errorf("update user: generate hash password: %w", err)
		slog.Error(err.Error())
		return nil, err
	}
	user := storage.User{
		Name:         userpb.GetName(),
		HashPassword: string(hashPassword),
		Email:        userpb.GetEmail(),
	}
	if err := s.authService.UpdateUser(ctx, &user); err != nil {
		err := fmt.Errorf("update user: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteUser(ctx context.Context, _ *grpcapi.DeleteUserRequest) (*emptypb.Empty, error) {
	userName := GetUserName(ctx)
	if userName == "" {
		panic("delete user: user name is empty")
	}
	if err := s.authService.DeleteUser(ctx, userName); err != nil {
		err := fmt.Errorf("delete user: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
