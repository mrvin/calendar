package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type UserDeleter struct {
	mock.Mock
}

func NewUserDeleter() *UserDeleter {
	return &UserDeleter{}
}

func (m *UserDeleter) DeleteUser(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	if len(args) == 0 {
		panic("no return value specified for DeleteUser")
	}

	return args.Error(0)
}
