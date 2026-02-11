package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Add pure Go Postgres driver for the database/sql package.
	_ "github.com/lib/pq"
	"github.com/mrvin/hw-otus-go/hw12-15calendar/pkg/retry"
)

const retriesConnect = 5

const maxOpenConns = 25
const maxIdleConns = 25
const connMaxLifetime = 5 // in minute

type Conf struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type Storage struct {
	db *sql.DB

	conf *Conf

	insertEvent *sql.Stmt
	getEvent    *sql.Stmt

	listEvents        *sql.Stmt
	listEventsForUser *sql.Stmt

	insertUser *sql.Stmt
	getUser    *sql.Stmt

	listUsers *sql.Stmt
}

func New(ctx context.Context, conf *Conf) (*Storage, error) {
	var st Storage

	st.conf = conf

	if err := st.RetryConnect(ctx, retriesConnect); err != nil {
		return nil, fmt.Errorf("new database connection: %w", err)
	}

	if err := st.prepareQuery(ctx); err != nil {
		return nil, fmt.Errorf("prepare query: %w", err)
	}

	return &st, nil
}

func (s *Storage) Connect(ctx context.Context) error {
	var err error
	dbConfStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		s.conf.Host, s.conf.Port, s.conf.User, s.conf.Password, s.conf.Name)
	s.db, err = sql.Open(s.conf.Driver, dbConfStr)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	// Setting db connections pool.
	s.db.SetMaxOpenConns(maxOpenConns)
	s.db.SetMaxIdleConns(maxIdleConns)
	s.db.SetConnMaxLifetime(connMaxLifetime * time.Minute)

	return nil
}

func (s *Storage) RetryConnect(ctx context.Context, retries int) error {
	retryConnect := retry.Retry(s.Connect, retries)
	if err := retryConnect(ctx); err != nil {
		return fmt.Errorf("connection db: %w", err)
	}

	return nil
}

func (s *Storage) prepareQuery(ctx context.Context) error {
	var err error
	fmtStrErr := "prepare \"%s\" query: %w"

	// User query prepare
	sqlInsertUser := `
		INSERT INTO users (
			name,
			hash_password,
			email
		)
		VALUES ($1, $2, $3)`
	s.insertUser, err = s.db.PrepareContext(ctx, sqlInsertUser)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "insertUser", err)
	}
	sqlGetUser := `
		SELECT name, hash_password, email
		FROM users
		WHERE name = $1`
	s.getUser, err = s.db.PrepareContext(ctx, sqlGetUser)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "getUser", err)
	}
	sqlGetAllUsers := `
		SELECT name, hash_password, email
		FROM users
	`
	s.listUsers, err = s.db.PrepareContext(ctx, sqlGetAllUsers)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "getAllUsers", err)
	}

	// Event query prepare
	sqlInsertEvent := `
		INSERT INTO events (
			title,
			description,
			start_time,
			stop_time,
			user_name
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	s.insertEvent, err = s.db.PrepareContext(ctx, sqlInsertEvent)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "insertEvent", err)
	}
	sqlGetEvent := `
		SELECT id, title, description, start_time, stop_time, user_name
		FROM events
		WHERE id = $1`
	s.getEvent, err = s.db.PrepareContext(ctx, sqlGetEvent)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "getEvent", err)
	}
	sqlGetAllEvent := `
		SELECT id, title, description, start_time, stop_time, user_name
		FROM events`
	s.listEvents, err = s.db.PrepareContext(ctx, sqlGetAllEvent)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "getAllEvents", err)
	}
	sqlGetEventsForUser := `
		SELECT id, title, description, start_time, stop_time, user_name
		FROM events
		WHERE user_name = $1`
	s.listEventsForUser, err = s.db.PrepareContext(ctx, sqlGetEventsForUser)
	if err != nil {
		return fmt.Errorf(fmtStrErr, "getEventsForUser", err)
	}

	return nil
}

func (s *Storage) Close() error {
	s.insertUser.Close()
	s.getUser.Close()
	s.listUsers.Close()

	s.insertEvent.Close()
	s.getEvent.Close()
	s.listEvents.Close()
	s.listEventsForUser.Close()

	return s.db.Close() //nolint:wrapcheck
}
