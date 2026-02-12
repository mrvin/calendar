package memorystorage

import (
	"context"
	"testing"

	"github.com/mrvin/calendar/internal/storage"
)

var ctx = context.Background()

func TestUserCRUD(t *testing.T) {
	st := New()

	storage.TestUserCRUD(ctx, t, st)
}

func TestEventCRUD(t *testing.T) {
	st := New()

	storage.TestEventCRUD(ctx, t, st)
}
