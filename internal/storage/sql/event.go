package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) (int64, error) {
	if err := s.insertEvent.QueryRowContext(ctx,
		event.Title,
		event.Description,
		event.StartTime,
		event.StopTime,
		event.UserName,
	).Scan(&event.ID); err != nil {
		return 0, fmt.Errorf("create event: %w", err)
	}

	return event.ID, nil
}

func (s *Storage) GetEvent(ctx context.Context, id int64) (*storage.Event, error) {
	var event storage.Event

	if err := s.getEvent.QueryRowContext(ctx, id).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&event.StopTime,
		&event.UserName,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: %d", storage.ErrNoEvent, id)
		}
		return nil, fmt.Errorf("can't get event with id: %d: %w", id, err)
	}

	return &event, nil
}

func (s *Storage) UpdateEvent(ctx context.Context, event *storage.Event) error {
	sqlUpdateEvent := `
		UPDATE events
		SET title = $2,
			description = $3,
			start_time = $4,
			stop_time = $5
		WHERE id = $1`
	res, err := s.db.ExecContext(ctx, sqlUpdateEvent,
		event.ID,
		event.Title,
		event.Description,
		event.StartTime,
		event.StopTime,
	)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("%w: %d", storage.ErrNoEvent, event.ID)
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, id int64) error {
	sqlDeleteEvent := "DELETE FROM events WHERE id = $1"
	res, err := s.db.ExecContext(ctx, sqlDeleteEvent, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("%w: %d", storage.ErrNoEvent, id)
	}

	return nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	events := make([]storage.Event, 0)

	rows, err := s.listEvents.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return events, nil
		}
		return nil, fmt.Errorf("can't get all events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var event storage.Event
		err = rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.StopTime,
			&event.UserName,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan next row: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return events, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

func (s *Storage) ListEventsForUser(ctx context.Context, name string) ([]storage.Event, error) {
	events := make([]storage.Event, 0)

	rows, err := s.listEventsForUser.QueryContext(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return events, nil
		}
		return nil, fmt.Errorf("can't get events for user with name: %s: %w", name, err)
	}
	defer rows.Close()

	for rows.Next() {
		var event storage.Event
		err = rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.StopTime,
			&event.UserName,
		)
		if err != nil {
			return nil, fmt.Errorf("can't scan next row: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return events, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}
