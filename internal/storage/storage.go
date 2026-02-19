package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")

	ErrDateBusy      = errors.New("date already busy")
	ErrEventNotFound = errors.New("event not found")
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, name string) (*User, error)
	DeleteUser(ctx context.Context, name string) error
}

type EventStorage interface {
	CreateEvent(ctx context.Context, event *Event) (uuid.UUID, error)
	GetEvent(ctx context.Context, username string, id uuid.UUID) (*Event, error)
	ListEvents(ctx context.Context, username string, start, end time.Time) ([]Event, error)
	UpdateEvent(ctx context.Context, username string, id uuid.UUID, event *Event) error
	DeleteEvent(ctx context.Context, username string, id uuid.UUID) error
}

type Storage interface {
	UserStorage
	EventStorage
}

type User struct {
	Name         string
	HashPassword string
	Email        string
	Role         string

	//	UpdatedAt   time.Time
	//	CreatedAt   time.Time
}

//nolint:tagliatelle
type Event struct {
	ID           uuid.UUID      `json:"id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	NotifyBefore *time.Duration `json:"notify_before"`
	Username     string         `json:"-"`

	//	UpdatedAt   time.Time
	//	CreatedAt   time.Time
}
