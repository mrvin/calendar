package memorystorage

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
		return fmt.Errorf("user with name %q already exists", user.Name)
	}

	s.mUsers[user.Name] = *user

	return nil
}

func (s *Storage) GetUser(_ context.Context, name string) (*storage.User, error) {
	s.muUsers.RLock()
	defer s.muUsers.RUnlock()

	user, ok := s.mUsers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", storage.ErrNoUser, name)
	}

	return &user, nil
}

func (s *Storage) UpdateUser(_ context.Context, newUser *storage.User) error {
	s.muUsers.Lock()
	defer s.muUsers.Unlock()

	oldUser, ok := s.mUsers[newUser.Name]
	if !ok {
		return fmt.Errorf("%w: %q", storage.ErrNoUser, newUser.Name)
	}

	oldUser.HashPassword = newUser.HashPassword
	oldUser.Email = newUser.Email

	s.mUsers[oldUser.Name] = oldUser

	return nil
}

func (s *Storage) DeleteUser(_ context.Context, name string) error {
	s.muUsers.Lock()
	user, ok := s.mUsers[name]
	if !ok {
		s.muUsers.Unlock()
		return fmt.Errorf("%w: %s", storage.ErrNoUser, name)
	}
	delete(s.mUsers, name)
	s.muUsers.Unlock()

	s.muEvents.Lock()
	for _, event := range s.mEvents {
		if event.UserName == user.Name {
			delete(s.mEvents, event.ID)
		}
	}
	s.muEvents.Unlock()

	return nil
}

func (s *Storage) ListUsers(_ context.Context) ([]storage.User, error) {
	users := make([]storage.User, 0)

	s.muUsers.RLock()
	for _, user := range s.mUsers {
		users = append(users, user)
	}
	s.muUsers.RUnlock()

	return users, nil
}
