package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrvin/calendar/internal/queue"
	"github.com/mrvin/calendar/internal/queue/rabbitmq"
	"github.com/mrvin/calendar/internal/storage"
)

type EventsLister interface {
	ListEventsToNotify(ctx context.Context, start, end time.Time) ([]storage.Event, error)
}

type App struct {
	lister      EventsLister
	qm          rabbitmq.Queue
	schedPeriod int
}

func New(lister EventsLister, qm rabbitmq.Queue, schedPeriod int) *App {
	return &App{
		lister:      lister,
		qm:          qm,
		schedPeriod: schedPeriod,
	}
}

func (a *App) Run(ctx context.Context) {
	ctx, _ = signal.NotifyContext(
		ctx,
		os.Interrupt,    // SIGINT, (Control-C)
		syscall.SIGTERM, // systemd
		syscall.SIGQUIT,
	)

	schedPeriod := time.Duration(a.schedPeriod) * time.Minute
	ticker := time.NewTicker(schedPeriod)
	for {
		select {
		case <-ticker.C:
			start := time.Now()
			events, err := a.lister.ListEventsToNotify(ctx, start, start.Add(schedPeriod))
			if err != nil {
				slog.Error("List events to notify", slog.String("error", err.Error()))
				continue
			}
			for _, event := range events {
				//TODO: generate new UUID
				alertEvent := queue.AlertEvent{
					ID:          event.ID,
					Title:       event.Title,
					Description: event.Description,
					StartTime:   event.StartTime,
					EndTime:     event.EndTime,
					Username:    event.Username,
				}
				byteAlertEvent, err := queue.EncodeAlertEvent(&alertEvent)
				if err != nil {
					slog.Error("Encode alert event", slog.String("error", err.Error()))
					continue
				}

				if err := a.qm.SendMsg(ctx, byteAlertEvent); err != nil {
					slog.Error("Send alert message", slog.String("error", err.Error()))
					continue
				}
				slog.Info("Put alert message in queue", slog.String("eventID", event.ID.String()))
			}
		case <-ctx.Done():
			ticker.Stop()
			slog.Info("Stop scheduler")
			return
		}
	}
}
