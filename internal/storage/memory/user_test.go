package memory

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mrvin/calendar/internal/storage"
)

func TestCreateUser_Success(t *testing.T) {
	s := New()
	ctx := context.Background()

	user := &storage.User{
		Name:         "testuser",
		HashPassword: "hash123",
		Email:        "test@example.com",
		Role:         "user",
	}

	err := s.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Verify user was created
	retrievedUser, err := s.GetUser(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrievedUser.Name != "testuser" || retrievedUser.Email != "test@example.com" {
		t.Errorf("User data mismatch: %+v", retrievedUser)
	}
}

func TestCreateUser_Duplicate(t *testing.T) {
	s := New()
	ctx := context.Background()

	user := &storage.User{
		Name:         "testuser",
		HashPassword: "hash123",
		Email:        "test@example.com",
	}

	err := s.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("First CreateUser failed: %v", err)
	}

	// Try to create duplicate
	err = s.CreateUser(ctx, user)
	if !errors.Is(err, storage.ErrUserExists) {
		t.Errorf("Expected ErrUserExists, got %v", err)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.GetUser(ctx, "nonexistent")
	if !errors.Is(err, storage.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	s := New()
	ctx := context.Background()

	user := &storage.User{Name: "testuser", HashPassword: "hash123"}
	s.CreateUser(ctx, user)

	err := s.DeleteUser(ctx, "testuser")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify user is deleted
	_, err = s.GetUser(ctx, "testuser")
	if !errors.Is(err, storage.ErrUserNotFound) {
		t.Errorf("User should be deleted")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	err := s.DeleteUser(ctx, "nonexistent")
	if !errors.Is(err, storage.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUser_DeletesUserEvents(t *testing.T) {
	s := New()
	ctx := context.Background()

	// Create user and event
	user := &storage.User{Name: "testuser", HashPassword: "hash123"}
	s.CreateUser(ctx, user)

	event := &storage.Event{
		Title:     "Test Event",
		Username:  "testuser",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	}
	eventID, err := s.CreateEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	// Delete user
	s.DeleteUser(ctx, "testuser")

	// Verify event is deleted
	_, err = s.GetEvent(ctx, "testuser", eventID)
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Event should be deleted when user is deleted")
	}
}

func TestConcurrentCreateUsers(t *testing.T) {
	s := New()
	ctx := context.Background()
	numGoroutines := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			user := &storage.User{
				Name:         "user" + string(rune(id)),
				HashPassword: "hash",
			}
			if err := s.CreateUser(ctx, user); err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Errorf("Concurrent CreateUser error: %v", err)
	}
}
