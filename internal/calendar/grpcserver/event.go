package grpcserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/storage"
	"github.com/mrvin/calendar/pkg/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateEvent(ctx context.Context, req *api.ReqCreateEvent) (*api.ResCreateEvent, error) {
	//TODO:add validation
	username, err := auth.GetUsernameFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting username from ctx: %v", err)
	}
	notifyBefore := req.GetNotifyBefore().AsDuration()
	//nolint:exhaustruct
	event := storage.Event{
		Title:        req.GetTitle(),
		Description:  req.GetDescription(),
		StartTime:    req.GetStartTime().AsTime(),
		EndTime:      req.GetEndTime().AsTime(),
		NotifyBefore: &notifyBefore,
		Username:     username,
	}

	id, err := s.storage.CreateEvent(ctx, &event)
	if err != nil {
		err = fmt.Errorf("saving event to storage: %w", err)
		if errors.Is(err, storage.ErrDateBusy) {
			return nil, status.Error(codes.Aborted, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &api.ResCreateEvent{Id: id.String()}, nil
}

func (s *Server) GetEvent(ctx context.Context, req *api.ReqGetEvent) (*api.ResEvent, error) {
	username, err := auth.GetUsernameFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting username from ctx: %v", err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse uuid: %v", err)
	}

	event, err := s.storage.GetEvent(ctx, username, id)
	if err != nil {
		err := fmt.Errorf("getting event from storage: %w", err)
		if errors.Is(err, storage.ErrEventNotFound) {
			return nil, status.Error(codes.NotFound, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &api.ResEvent{
		Id:           event.ID.String(),
		Title:        event.Title,
		Description:  event.Description,
		StartTime:    timestamppb.New(event.StartTime),
		EndTime:      timestamppb.New(event.EndTime),
		NotifyBefore: durationpb.New(*event.NotifyBefore),
	}, nil
}

func (s *Server) ListEvents(ctx context.Context, req *api.ReqListEvents) (*api.ResListEvents, error) {
	//TODO: validet CheckValid

	username, err := auth.GetUsernameFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting username from ctx: %v", err)
	}

	events, err := s.storage.ListEvents(ctx, username, req.GetStartTime().AsTime(), req.GetEndTime().AsTime())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting list events from storage: %v", err)
	}

	pbEvents := make([]*api.ResEvent, len(events))
	for i, event := range events {
		pbEvents[i] = &api.ResEvent{
			Id:           event.ID.String(),
			Title:        event.Title,
			Description:  event.Description,
			StartTime:    timestamppb.New(event.StartTime),
			EndTime:      timestamppb.New(event.EndTime),
			NotifyBefore: durationpb.New(*event.NotifyBefore),
		}
	}

	return &api.ResListEvents{Events: pbEvents}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *api.ReqUpdateEvent) (*emptypb.Empty, error) {
	//TODO: validet CheckValid
	username, err := auth.GetUsernameFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting username from ctx: %v", err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse uuid: %v", err)
	}

	notifyBefore := req.GetNotifyBefore().AsDuration()
	//nolint:exhaustruct
	event := storage.Event{
		Title:        req.GetTitle(),
		Description:  req.GetDescription(),
		StartTime:    req.GetStartTime().AsTime(),
		EndTime:      req.GetEndTime().AsTime(),
		NotifyBefore: &notifyBefore,
	}

	if err := s.storage.UpdateEvent(ctx, username, id, &event); err != nil {
		err = fmt.Errorf("updating event to storage: %w", err)
		if errors.Is(err, storage.ErrDateBusy) {
			return nil, status.Error(codes.Aborted, err.Error()) //nolint:wrapcheck
		}
		if errors.Is(err, storage.ErrEventNotFound) {
			return nil, status.Error(codes.NotFound, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *api.ReqDeleteEvent) (*emptypb.Empty, error) {
	username, err := auth.GetUsernameFromCtx(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "getting username from ctx: %v", err)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "parse uuid: %v", err)
	}

	if err := s.storage.DeleteEvent(ctx, username, id); err != nil {
		err := fmt.Errorf("deleting event from storage: %w", err)
		if errors.Is(err, storage.ErrEventNotFound) {
			return nil, status.Error(codes.NotFound, err.Error()) //nolint:wrapcheck
		}
		return nil, status.Error(codes.Internal, err.Error()) //nolint:wrapcheck
	}

	return &emptypb.Empty{}, nil
}
