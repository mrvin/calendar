package grpcserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	"github.com/mrvin/calendar/pkg/api"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) Register(ctx context.Context, req *api.ReqRegister) (*emptypb.Empty, error) {
	ctx = logger.WithUsername(ctx, req.GetUsername())
	//TODO:add validation
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate hash password: %v", err)
	}
	user := storage.User{
		Name:         req.GetUsername(),
		HashPassword: string(hashPassword),
		Email:        req.GetEmail(),
		Role:         "user",
	}
	if err := s.storage.CreateUser(ctx, &user); err != nil {
		err = fmt.Errorf("saving user to storage: %w", err)
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.Aborted, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) Login(ctx context.Context, req *api.ReqLogin) (*api.ResLogin, error) {
	ctx = logger.WithUsername(ctx, req.GetUsername())
	tokenString, err := s.auth.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &api.ResLogin{AccessToken: tokenString}, nil
}

func (s *Server) GetUser(ctx context.Context, _ *emptypb.Empty) (*api.ResUser, error) {
	username, err := auth.GetUsernameFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting username from ctx: %v", err)
	}

	user, err := s.storage.GetUser(ctx, username)
	if err != nil {
		err := fmt.Errorf("getting user from storage: %w", err)
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &api.ResUser{
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *Server) DeleteUser(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	username, err := auth.GetUsernameFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting username from ctx: %v", err)
	}

	if err := s.storage.DeleteUser(ctx, username); err != nil {
		err := fmt.Errorf("deleting user from storage: %w", err)
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &emptypb.Empty{}, nil
}
