package grpcserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/grpcapi"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateEvent(ctx context.Context, pbEvent *grpcapi.CreateEventRequest) (*grpcapi.CreateEventResponse, error) {
	if err := pbEvent.GetStartTime().CheckValid(); err != nil {
		err = fmt.Errorf("incorrect value StartTime: %w", err)
		slog.Error(err.Error())
		return nil, err
	}
	if err := pbEvent.GetStopTime().CheckValid(); err != nil {
		err = fmt.Errorf("incorrect value StopTime: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	userName := GetUserName(ctx)
	user, err := s.authService.GetUser(ctx, userName)
	if err != nil {
		err = fmt.Errorf("get user: %w", err)
		slog.Error(err.Error())
		return nil, err
	}
	event := storage.Event{
		ID:          0,
		Title:       pbEvent.GetTitle(),
		Description: pbEvent.GetDescription(),
		StartTime:   pbEvent.GetStartTime().AsTime(),
		StopTime:    pbEvent.GetStopTime().AsTime(),
		UserName:    user.Name,
	}

	id, err := s.eventService.CreateEvent(ctx, &event)
	if err != nil {
		err = fmt.Errorf("create event: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &grpcapi.CreateEventResponse{Id: id}, nil
}

func (s *Server) Login(ctx context.Context, req *grpcapi.LoginRequest) (*grpcapi.LoginResponse, error) {
	tokenString, err := s.authService.Authenticate(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	return &grpcapi.LoginResponse{AccessToken: tokenString}, nil
}

func (s *Server) GetEvent(ctx context.Context, req *grpcapi.GetEventRequest) (*grpcapi.EventResponse, error) {
	event, err := s.eventService.GetEvent(ctx, req.GetId())
	if err != nil {
		err := fmt.Errorf("get event: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &grpcapi.EventResponse{
		Id:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		StartTime:   timestamppb.New(event.StartTime),
		StopTime:    timestamppb.New(event.StopTime),
		UserName:    event.UserName,
	}, nil
}

func (s *Server) ListEventsForUser(ctx context.Context, req *grpcapi.ListEventsForUserRequest) (*grpcapi.ListEventsResponse, error) {
	if err := req.GetDate().CheckValid(); err != nil {
		return nil, fmt.Errorf("incorrect value date: %w", err)
	}
	date := req.GetDate().AsTime()

	userName := GetUserName(ctx)
	if userName == "" {
		err := fmt.Errorf("user name is empty")
		slog.Error(err.Error())
		return nil, err
	}

	events, err := s.eventService.ListEventsForUser(ctx, userName, date, int(req.GetDays()))
	if err != nil {
		err := fmt.Errorf("get events for user: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	pbEvents := make([]*grpcapi.EventResponse, len(events), len(events))
	for i, event := range events {
		pbEvents[i] = &grpcapi.EventResponse{
			Id:          event.ID,
			Title:       event.Title,
			Description: event.Description,
			StartTime:   timestamppb.New(event.StartTime),
			StopTime:    timestamppb.New(event.StopTime),
			UserName:    event.UserName,
		}
	}

	return &grpcapi.ListEventsResponse{Events: pbEvents}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, pbEvent *grpcapi.UpdateEventRequest) (*emptypb.Empty, error) {
	if err := pbEvent.GetStartTime().CheckValid(); err != nil {
		err = fmt.Errorf("incorrect value StartTime: %w", err)
		slog.Error(err.Error())
		return nil, err
	}
	if err := pbEvent.GetStopTime().CheckValid(); err != nil {
		err = fmt.Errorf("incorrect value StopTime: %w", err)
		slog.Error(err.Error())
		return nil, err
	}
	event := storage.Event{
		ID:          0,
		Title:       pbEvent.GetTitle(),
		Description: pbEvent.GetDescription(),
		StartTime:   pbEvent.GetStartTime().AsTime(),
		StopTime:    pbEvent.GetStopTime().AsTime(),
		UserName:    pbEvent.GetUserName(),
	}

	if err := s.eventService.UpdateEvent(ctx, &event); err != nil {
		err := fmt.Errorf("update event: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *grpcapi.DeleteEventRequest) (*emptypb.Empty, error) {
	if err := s.eventService.DeleteEvent(ctx, req.GetId()); err != nil {
		err := fmt.Errorf("delete event: %w", err)
		slog.Error(err.Error())
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
