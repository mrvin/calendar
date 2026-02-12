package sqlstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/mrvin/calendar/internal/storage"
)

func (s *Storage) CreateUser(ctx context.Context, user *storage.User) error {
	if _, err := s.insertUser.ExecContext(ctx, user.Name, user.HashPassword, user.Email); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) {
			if pgErr.Code.Name() == "unique_violation" {
				return fmt.Errorf("db: %w: %q", storage.ErrUserExists, user.Name)
			}
		}
		return fmt.Errorf("db: %w", err)
	}

	return nil
}

func (s *Storage) GetUser(ctx context.Context, name string) (*storage.User, error) {
	var user storage.User

	if err := s.getUser.QueryRowContext(ctx, name).Scan(&user.Name, &user.HashPassword, &user.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", storage.ErrNoUser, name)
		}
		return nil, fmt.Errorf("can't scan user with name: %s: %w", name, err)
	}

	return &user, nil
}

func (s *Storage) UpdateUser(ctx context.Context, user *storage.User) error {
	res, err := s.db.ExecContext(ctx, "update users set email = $3 where name = $1", user.Name, user.Email)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("%w: %q", storage.ErrNoUser, user.Name)
	}

	return nil
}

func (s *Storage) DeleteUser(ctx context.Context, name string) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE name = $1", name)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("%w: %s", storage.ErrNoUser, name)
	}

	return nil
}

func (s *Storage) ListUsers(ctx context.Context) ([]storage.User, error) {
	users := make([]storage.User, 0)

	rows, err := s.listUsers.QueryContext(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return users, nil
		}
		return nil, fmt.Errorf("can't get all users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user storage.User
		err = rows.Scan(&user.Name, &user.HashPassword, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("can't scan next row: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return users, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
