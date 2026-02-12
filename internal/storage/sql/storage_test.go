package sqlstorage

import (
	"context"
	"testing"

	"github.com/mrvin/calendar/internal/storage"
)

var ctx = context.Background()

var confDBTest = Conf{"postgres", "postgres", 5432, "event-db", "event-db", "event-db"}

func TestUserCRUD(t *testing.T) {
	st, err := New(ctx, &confDBTest)
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer st.Close()

	storage.TestUserCRUD(ctx, t, st)
}

func TestEventCRUD(t *testing.T) {
	st, err := New(ctx, &confDBTest)
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer st.Close()

	storage.TestEventCRUD(ctx, t, st)
}
