package storage

import (
	"context"
	"errors"
	"time"
)

var ErrNoUser = errors.New("no user with name")
var ErrUserExists = errors.New("user exists")
var ErrNoEvent = errors.New("no event with id")

type EventStorage interface {
	CreateEvent(ctx context.Context, event *Event) (int64, error)
	GetEvent(ctx context.Context, id int64) (*Event, error)
	UpdateEvent(ctx context.Context, event *Event) error
	DeleteEvent(ctx context.Context, id int64) error

	ListEvents(ctx context.Context) ([]Event, error)
	ListEventsForUser(ctx context.Context, name string) ([]Event, error)
}

type UserStorage interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, name string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, name string) error

	ListUsers(ctx context.Context) ([]User, error)
}

type Storage interface {
	EventStorage
	UserStorage
}

//nolint:tagliatelle
type Event struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time"`
	StopTime    time.Time `json:"stop_time,omitempty"`
	UserName    string    `json:"user_name"`
	//	CreatedAt   time.Time
	//	UpdatedAt   time.Time
}

//nolint:tagliatelle
type User struct {
	Name         string `json:"name"`
	HashPassword string `json:"hash_password"`
	Email        string `json:"email"`
	Role         string `json:"role"`
}
