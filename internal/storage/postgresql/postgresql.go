package postgresql

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mrvin/calendar/pkg/retry"
)

const (
	retriesPing = 5

	maxOpenConns    = 25
	minOpenConns    = 5
	maxConnIdleTime = 30 // in minute
	maxConnLifetime = 60 // in minute
)

type Conf struct {
	Host     string `env:"POSTGRES_HOST"     yaml:"host"`
	Port     string `env:"POSTGRES_PORT"     yaml:"port"`
	User     string `env:"POSTGRES_USER"     yaml:"user"`
	Password string `env:"POSTGRES_PASSWORD" yaml:"password"`
	Name     string `env:"POSTGRES_DB"       yaml:"name"`
}

type Storage struct {
	db *pgxpool.Pool

	conf *Conf
}

func New(ctx context.Context, conf *Conf) (*Storage, error) {
	var st Storage

	st.conf = conf

	if err := st.connect(ctx); err != nil {
		return nil, err
	}

	return &st, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) connect(ctx context.Context) error {
	dbConfStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=disable",
		s.conf.User,
		s.conf.Password,
		net.JoinHostPort(s.conf.Host, s.conf.Port),
		s.conf.Name,
	)
	config, err := pgxpool.ParseConfig(dbConfStr)
	if err != nil {
		return fmt.Errorf("parsing config string: %w", err)
	}
	config.MaxConns = maxOpenConns
	config.MinConns = minOpenConns
	config.MaxConnIdleTime = time.Minute * maxConnIdleTime
	config.MaxConnLifetime = time.Minute * maxConnLifetime

	s.db, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("create connection pool: %w", err)
	}

	retryPing := retry.Retry(s.db.Ping, retriesPing)
	if err := retryPing(ctx); err != nil {
		return fmt.Errorf("connection db: %w", err)
	}

	return nil
}
