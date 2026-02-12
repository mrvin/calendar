package eventservice

import (
	"context"
	"errors"
	"time"

	"github.com/mrvin/calendar/internal/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var ErrStopTimeBeforeStartTime = errors.New("event ends before starts")

type EventService struct {
	eventSt storage.EventStorage
	tr      trace.Tracer // Think about it.
}

func New(st storage.EventStorage) *EventService {
	return &EventService{
		st,
		otel.GetTracerProvider().Tracer("Event service"),
	}
}

func (e *EventService) CreateEvent(ctx context.Context, event *storage.Event) (int64, error) {
	if event.StopTime.Before(event.StartTime) {
		return 0, ErrStopTimeBeforeStartTime
	}

	cctx, sp := e.tr.Start(ctx, "CreateEvent")
	defer sp.End()

	return e.eventSt.CreateEvent(cctx, event)
}

func (e *EventService) GetEvent(ctx context.Context, id int64) (*storage.Event, error) {
	cctx, sp := e.tr.Start(ctx, "GetEvent")
	defer sp.End()

	return e.eventSt.GetEvent(cctx, id)
}

func (e *EventService) UpdateEvent(ctx context.Context, event *storage.Event) error {
	cctx, sp := e.tr.Start(ctx, "UpdateEvent")
	defer sp.End()

	return e.eventSt.UpdateEvent(cctx, event)
}

func (e *EventService) DeleteEvent(ctx context.Context, id int64) error {
	cctx, sp := e.tr.Start(ctx, "DeleteEvent")
	defer sp.End()

	return e.eventSt.DeleteEvent(cctx, id)
}

/*
func (a *App) ListEvents(ctx context.Context) ([]storage.Event, error) {
	return a.storage.GetAllEvents(ctx)
}
*/

// TODO: implement at the database level.
func (e *EventService) ListEventsForUser(ctx context.Context, name string, startPeriod time.Time, days int) ([]storage.Event, error) {
	cctx, sp := e.tr.Start(ctx, "GetEventsForUser")
	defer sp.End()

	events, err := e.eventSt.ListEventsForUser(cctx, name)
	if err != nil {
		return nil, err
	}
	if days == 0 {
		return events, nil
	}
	stopPeriod := startPeriod.AddDate(0, 0, days)
	var eventsFromInterval []storage.Event
	for _, event := range events {
		if event.StartTime.After(startPeriod) && event.StartTime.Before(stopPeriod) ||
			event.StopTime.After(startPeriod) && event.StopTime.Before(stopPeriod) {
			eventsFromInterval = append(eventsFromInterval, event)
		}
	}

	return eventsFromInterval, nil
}
