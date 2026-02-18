package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/storage"
)

func (s *Storage) CreateEvent(_ context.Context, event *storage.Event) (uuid.UUID, error) {
	event.ID = uuid.New()

	s.muEvents.Lock()
	defer s.muEvents.Unlock()

	for _, existEvent := range s.mEvents {
		if existEvent.Username == event.Username {
			if existEvent.StartTime.Before(event.EndTime) && existEvent.EndTime.After(event.StartTime) {
				return uuid.Nil, storage.ErrDateBusy
			}
		}
	}
	s.mEvents[event.ID] = *event

	return event.ID, nil
}

func (s *Storage) GetEvent(_ context.Context, username string, id uuid.UUID) (*storage.Event, error) {
	s.muEvents.RLock()
	event, ok := s.mEvents[id]
	s.muEvents.RUnlock()
	if !ok || event.Username != username {
		return nil, fmt.Errorf("%w: %s", storage.ErrEventNotFound, id)
	}

	return &event, nil
}

func (s *Storage) DeleteEvent(_ context.Context, username string, id uuid.UUID) error {
	s.muEvents.Lock()
	defer s.muEvents.Unlock()

	if event, ok := s.mEvents[id]; !ok || event.Username != username {
		return fmt.Errorf("%w: %s", storage.ErrEventNotFound, id)
	}
	delete(s.mEvents, id)

	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, username string, id uuid.UUID, event *storage.Event) error {
	s.muEvents.Lock()
	defer s.muEvents.Unlock()

	oldEvent, ok := s.mEvents[id]
	if !ok || oldEvent.Username != username {
		return fmt.Errorf("%w: %s", storage.ErrEventNotFound, id)
	}

	for eventID, existEvent := range s.mEvents {
		if existEvent.Username == username && eventID != id {
			if existEvent.StartTime.Before(event.EndTime) && existEvent.EndTime.After(event.StartTime) {
				return storage.ErrDateBusy
			}
		}
	}

	s.mEvents[id] = *event

	return nil
}

func (s *Storage) ListEvents(_ context.Context, username string, start, end time.Time) ([]storage.Event, error) {
	events := make([]storage.Event, 0)

	s.muEvents.RLock()
	for _, event := range s.mEvents {
		if event.Username == username {
			if event.StartTime.Before(end) && event.EndTime.After(start) {
				events = append(events, event)
			}
		}
	}
	s.muEvents.RUnlock()

	return events, nil
}
