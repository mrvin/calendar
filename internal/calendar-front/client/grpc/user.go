package grpcclient

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/grpcapi"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

func (c *Client) Registration(ctx context.Context, name, password, email string) error {
	user := &grpcapi.CreateUserRequest{
		Name:     name,
		Password: password,
		Email:    email,
	}
	_, err := c.userService.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("gRPC: %w", err)
	}
	slog.Debug("Added user",
		slog.String("username", user.GetName()),
		slog.String("password", user.GetPassword()),
		slog.String("email", user.GetEmail()),
	)

	return nil
}

func (c *Client) Login(ctx context.Context, name, password string) (string, error) {
	reqLogin := &grpcapi.LoginRequest{
		Username: name,
		Password: password,
	}
	response, err := c.userService.Login(ctx, reqLogin)
	if err != nil {
		return "", fmt.Errorf("gRPC: %w", err)
	}

	return response.GetAccessToken(), nil
}

func (c *Client) GetUser(ctx context.Context, token string) (*storage.User, error) {
	reqUser := &grpcapi.GetUserRequest{
		AccessToken: token,
	}
	user, err := c.userService.GetUser(ctx, reqUser)
	if err != nil {
		return nil, fmt.Errorf("gRPC: %w", err)
	}

	return &storage.User{
		Name:         user.GetName(),
		HashPassword: user.GetHashPassword(),
		Email:        user.GetEmail(),
	}, nil
}

func (c *Client) UpdateUser(ctx context.Context, token, name, password, email string) error {
	user := &grpcapi.UpdateUserRequest{
		AccessToken: token,
		Name:        name,
		Password:    password,
		Email:       email,
	}
	_, err := c.userService.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("gRPC: %w", err)
	}
	slog.Debug("Update user",
		slog.String("username", name),
	)

	return nil
}

func (c *Client) DeleteUser(ctx context.Context, token string) error {
	reqDeleteUser := &grpcapi.DeleteUserRequest{
		AccessToken: token,
	}
	if _, err := c.userService.DeleteUser(ctx, reqDeleteUser); err != nil {
		return fmt.Errorf("gRPC: %w", err)
	}

	return nil
}

func (c *Client) ListUsers(ctx context.Context, token string) ([]storage.User, error) {
	reqListUsers := &grpcapi.ListUsersRequest{
		AccessToken: token,
	}
	users, err := c.userService.ListUsers(ctx, reqListUsers)
	if err != nil {
		return nil, fmt.Errorf("gRPC: %w", err)
	}

	listUsers := make([]storage.User, 0, len(users.GetUsers()))
	for _, user := range users.GetUsers() {
		listUsers = append(listUsers, storage.User{
			Name:         user.GetName(),
			HashPassword: user.GetHashPassword(),
			Email:        user.GetEmail(),
		})
	}

	return listUsers, nil
}
