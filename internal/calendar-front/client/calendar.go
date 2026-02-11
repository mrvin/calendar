package client

import (
	"context"
	"time"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

type Calendar interface {
	Registration(ctx context.Context, name, password, email string) error
	Login(ctx context.Context, name, password string) (string, error)

	GetUser(ctx context.Context, token string) (*storage.User, error)
	UpdateUser(ctx context.Context, token, name, password, email string) error
	DeleteUser(ctx context.Context, token string) error

	ListUsers(ctx context.Context, token string) ([]storage.User, error)

	CreateEvent(ctx context.Context, token string,
		title, description string,
		startTime, stopTime time.Time) (int64, error)
	GetEvent(ctx context.Context, token string, id int64) (*storage.Event, error)
	UpdateEvent(ctx context.Context, token string,
		title, description string,
		startTime, stopTime time.Time,
	) error
	DeleteEvent(ctx context.Context, token string, id int64) error

	// ListEvents(ctx context.Context) ([]Event, error)
	ListEventsForUser(ctx context.Context, token string, days int) ([]storage.Event, error)
}
