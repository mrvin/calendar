package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mrvin/calendar/internal/storage"
)

func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) (uuid.UUID, error) {
	sqlInsertEvent := `
		INSERT INTO events (
			title,
			description,
			start_time,
			end_time,
			notify_before,
			username
		)
		SELECT $1, $2, $3, $4, $5, $6
        WHERE NOT EXISTS (
            SELECT 1
            FROM events e
            WHERE e.username = $6
              AND e.start_time < $4
              AND e.end_time > $3
        )
		RETURNING id`
	if err := s.db.QueryRow(ctx, sqlInsertEvent,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.NotifyBefore,
		event.Username,
	).Scan(&event.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, storage.ErrDateBusy
		}
		return uuid.Nil, fmt.Errorf("insert event: %w", err)
	}

	return event.ID, nil
}

func (s *Storage) GetEvent(ctx context.Context, username string, id uuid.UUID) (*storage.Event, error) {
	sqlGetEvent := `
		SELECT id, title, description, start_time, end_time, notify_before, username
		FROM events
		WHERE username = $1 AND id = $2`
	var event storage.Event
	if err := s.db.QueryRow(ctx, sqlGetEvent, username, id).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&event.EndTime,
		&event.NotifyBefore,
		&event.Username,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get event: %w: %q", storage.ErrEventNotFound, id)
		}
		return nil, fmt.Errorf("get event: %q: %w", id, err)
	}

	return &event, nil
}

func (s *Storage) UpdateEvent(ctx context.Context, username string, id uuid.UUID, event *storage.Event) error {
	sqlUpdateEvent := `
		UPDATE events
		SET title = $1,
		    description = $2,
		    start_time = $3,
		    end_time = $4,
		    notify_before = $5
		WHERE username = $6
		  AND id = $7
		  AND NOT EXISTS (
		      SELECT 1
		      FROM events e
		      WHERE e.username = $6
		        AND e.id != $7
		        AND e.start_time < $4
		        AND e.end_time > $3
		  )`
	res, err := s.db.Exec(ctx, sqlUpdateEvent,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.NotifyBefore,
		username,
		id,
	)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}

	if res.RowsAffected() == 0 {
		sqlCheckEvent := "SELECT 1 FROM events WHERE username = $1 AND id = $2"
		if err := s.db.QueryRow(ctx, sqlCheckEvent, username, id).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("update event: %q: %w", id, storage.ErrEventNotFound)
			}
			return fmt.Errorf("update event: %w", err)
		}
		return storage.ErrDateBusy
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, username string, id uuid.UUID) error {
	sqlDeleteEvent := "DELETE FROM events WHERE username = $1 AND id = $2"
	res, err := s.db.Exec(ctx, sqlDeleteEvent, username, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("delete event: %q: %w", id, storage.ErrEventNotFound)
	}

	return nil
}

func (s *Storage) ListEvents(ctx context.Context, username string, start, end time.Time) ([]storage.Event, error) {
	sqlListEvents := `
		SELECT id, title, description, start_time, end_time, notify_before, username
		FROM events
		WHERE username = $1 AND start_time < $3 AND end_time > $2
		ORDER BY start_time DESC`
	rows, err := s.db.Query(ctx, sqlListEvents, username, start, end)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []storage.Event{}, nil
		}
		return nil, fmt.Errorf("list events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[storage.Event])
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	return events, nil
}
