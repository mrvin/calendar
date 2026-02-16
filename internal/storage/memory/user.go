package memory

import (
	"context"
	"fmt"

	"github.com/mrvin/calendar/internal/storage"
)

func (s *Storage) CreateUser(_ context.Context, user *storage.User) error {
	s.muUsers.Lock()
	defer s.muUsers.Unlock()

	_, ok := s.mUsers[user.Name]
	if ok {
		return fmt.Errorf("%w: %q", storage.ErrUserExists, user.Name)
	}
	s.mUsers[user.Name] = *user

	return nil
}

func (s *Storage) GetUser(_ context.Context, name string) (*storage.User, error) {
	s.muUsers.RLock()
	defer s.muUsers.RUnlock()

	user, ok := s.mUsers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", storage.ErrUserNotFound, name)
	}

	return &user, nil
}

func (s *Storage) DeleteUser(_ context.Context, name string) error {
	s.muUsers.Lock()
	defer s.muUsers.Unlock()

	if _, ok := s.mUsers[name]; !ok {
		return fmt.Errorf("%w: %q", storage.ErrUserNotFound, name)
	}
	s.muEvents.Lock()
	for id, event := range s.mEvents {
		if event.Username == name {
			delete(s.mEvents, id)
		}
	}
	s.muEvents.Unlock()

	delete(s.mUsers, name)

	return nil
}
