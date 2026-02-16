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

	ErrEventNotFound = errors.New("event not found")
	ErrDateBusy      = errors.New("date already busy")
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, name string) (*User, error)
	DeleteUser(ctx context.Context, name string) error
}

type EventStorage interface {
	CreateEvent(ctx context.Context, event *Event) (uuid.UUID, error)
	GetEvent(ctx context.Context, username string, id uuid.UUID) (*Event, error)
	ListEvents(ctx context.Context, username string, startWindow, endWindow time.Time) ([]Event, error)
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

type Event struct {
	ID           uuid.UUID
	Title        string
	Description  string
	StartTime    time.Time
	EndTime      time.Time
	NotifyBefore *time.Duration
	Username     string

	//	UpdatedAt   time.Time
	//	CreatedAt   time.Time
}
