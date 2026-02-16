package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mrvin/calendar/internal/storage"
)

func (s *Storage) CreateUser(ctx context.Context, user *storage.User) error {
	sqlInsertUser := `
		INSERT INTO users (
			name,
			hash_password,
			email,
			role
		)
		VALUES ($1, $2, $3, $4)`
	if _, err := s.db.Exec(ctx, sqlInsertUser,
		user.Name,
		user.HashPassword,
		user.Email,
		user.Role,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 = unique_violation
			return fmt.Errorf("insert user: %w: %q", storage.ErrUserExists, user.Name)
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, name string) (*storage.User, error) {
	sqlGetUser := `
		SELECT name, hash_password, email, role
		FROM users
		WHERE name = $1`
	var user storage.User
	if err := s.db.QueryRow(ctx, sqlGetUser, name).Scan(
		&user.Name,
		&user.HashPassword,
		&user.Email,
		&user.Role,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get user: %w: %q", storage.ErrUserNotFound, name)
		}
		return nil, fmt.Errorf("get user: %q: %w", name, err)
	}

	return &user, nil
}

func (s *Storage) DeleteUser(ctx context.Context, name string) error {
	sqlDeleteUser := `DELETE FROM users WHERE name = $1`
	res, err := s.db.Exec(ctx, sqlDeleteUser, name)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("delete user: %q: %w", name, storage.ErrUserExists)
	}

	return nil
}
