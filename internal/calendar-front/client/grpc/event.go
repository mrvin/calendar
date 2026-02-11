package grpcclient

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/grpcapi"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *Client) CreateEvent(
	ctx context.Context, token, title, description string,
	startTime, stopTime time.Time,
) (int64, error) {
	event := &grpcapi.CreateEventRequest{
		AccessToken: token,
		Title:       title,
		Description: description,
		StartTime:   timestamppb.New(startTime),
		StopTime:    timestamppb.New(stopTime),
	}
	response, err := c.eventService.CreateEvent(ctx, event)
	if err != nil {
		return 0, fmt.Errorf("gRPC: %w", err)
	}
	slog.Debug("Added event",
		slog.Int64("id", response.GetId()),
		slog.String("title", event.GetTitle()),
		slog.String("Description", event.GetDescription()),
	)

	return response.GetId(), nil
}

func (c *Client) GetEvent(ctx context.Context, token string, id int64) (*storage.Event, error) {
	reqEvent := &grpcapi.GetEventRequest{
		AccessToken: token,
		Id:          id,
	}
	event, err := c.eventService.GetEvent(ctx, reqEvent)
	if err != nil {
		return nil, fmt.Errorf("gRPC: %w", err)
	}

	return &storage.Event{
		ID:          event.GetId(),
		Title:       event.GetTitle(),
		Description: event.GetDescription(),
		StartTime:   event.GetStartTime().AsTime(),
		StopTime:    event.GetStopTime().AsTime(),
	}, nil
}

func (c *Client) UpdateEvent(
	ctx context.Context, token, title, description string,
	startTime, stopTime time.Time,
) error {
	event := &grpcapi.UpdateEventRequest{
		AccessToken: token,
		Title:       title,
		Description: description,
		StartTime:   timestamppb.New(startTime),
		StopTime:    timestamppb.New(stopTime),
	}
	if _, err := c.eventService.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("gRPC: %w", err)
	}

	return nil
}

func (c *Client) DeleteEvent(ctx context.Context, token string, id int64) error {
	reqEvent := &grpcapi.DeleteEventRequest{
		AccessToken: token,
		Id:          id,
	}
	if _, err := c.eventService.DeleteEvent(ctx, reqEvent); err != nil {
		return fmt.Errorf("gRPC: %w", err)
	}

	return nil
}

func (c *Client) ListEventsForUser(ctx context.Context, token string, days int) ([]storage.Event, error) {
	reqUser := &grpcapi.ListEventsForUserRequest{
		AccessToken: token,
		Days:        int32(days),
		Date:        timestamppb.New(time.Now()),
	}

	events, err := c.eventService.ListEventsForUser(ctx, reqUser)
	if err != nil {
		return nil, fmt.Errorf("gRPC: %w", err)
	}
	fmt.Printf("Events: %v\n", events)
	listEvents := make([]storage.Event, 0, len(events.GetEvents()))
	for _, event := range events.GetEvents() {
		listEvents = append(listEvents, storage.Event{
			ID:          event.GetId(),
			Title:       event.GetTitle(),
			Description: event.GetDescription(),
			StartTime:   event.GetStartTime().AsTime(),
			StopTime:    event.GetStopTime().AsTime(),
		})
	}

	return listEvents, nil
}
