package memory

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/storage"
)

func TestCreateEvent_Success(t *testing.T) {
	s := New()
	ctx := context.Background()

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

	if eventID == uuid.Nil {
		t.Errorf("Expected valid event ID")
	}
}

func TestCreateEvent_TimingConflict(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()

	// Create first event
	event1 := &storage.Event{
		Title:     "Event 1",
		Username:  "testuser",
		StartTime: now,
		EndTime:   now.Add(2 * time.Hour),
	}
	_, err := s.CreateEvent(ctx, event1)
	if err != nil {
		t.Fatalf("First CreateEvent failed: %v", err)
	}

	// Try to create overlapping event
	event2 := &storage.Event{
		Title:     "Event 2",
		Username:  "testuser",
		StartTime: now.Add(1 * time.Hour),
		EndTime:   now.Add(3 * time.Hour),
	}
	_, err = s.CreateEvent(ctx, event2)
	if !errors.Is(err, storage.ErrDateBusy) {
		t.Errorf("Expected ErrDateBusy, got %v", err)
	}
}

func TestCreateEvent_DifferentUsersCanOverlap(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()

	// Create event for user1
	event1 := &storage.Event{
		Title:     "Event 1",
		Username:  "user1",
		StartTime: now,
		EndTime:   now.Add(2 * time.Hour),
	}
	_, err := s.CreateEvent(ctx, event1)
	if err != nil {
		t.Fatalf("CreateEvent for user1 failed: %v", err)
	}

	// Create overlapping event for user2 (should succeed)
	event2 := &storage.Event{
		Title:     "Event 2",
		Username:  "user2",
		StartTime: now,
		EndTime:   now.Add(2 * time.Hour),
	}
	_, err = s.CreateEvent(ctx, event2)
	if err != nil {
		t.Errorf("Different users should be able to have overlapping events, got %v", err)
	}
}

func TestGetEvent_Success(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	event := &storage.Event{
		Title:     "Test Event",
		Username:  "testuser",
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}
	eventID, _ := s.CreateEvent(ctx, event)

	retrieved, err := s.GetEvent(ctx, "testuser", eventID)
	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}

	if retrieved.Title != "Test Event" {
		t.Errorf("Event title mismatch")
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	_, err := s.GetEvent(ctx, "testuser", uuid.New())
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestGetEvent_WrongUser(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := &storage.Event{
		Title:     "Test Event",
		Username:  "user1",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	}
	eventID, _ := s.CreateEvent(ctx, event)

	// Try to get event as different user
	_, err := s.GetEvent(ctx, "user2", eventID)
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Should not allow different user to get event")
	}
}

func TestUpdateEvent_Success(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	event := &storage.Event{
		Title:     "Original Title",
		Username:  "testuser",
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}
	eventID, _ := s.CreateEvent(ctx, event)

	// Update event
	updated := &storage.Event{
		Title:     "Updated Title",
		Username:  "testuser",
		StartTime: now.Add(2 * time.Hour),
		EndTime:   now.Add(3 * time.Hour),
	}
	err := s.UpdateEvent(ctx, "testuser", eventID, updated)
	if err != nil {
		t.Fatalf("UpdateEvent failed: %v", err)
	}

	// Verify update
	retrieved, _ := s.GetEvent(ctx, "testuser", eventID)
	if retrieved.Title != "Updated Title" {
		t.Errorf("Event title not updated")
	}
	if !retrieved.StartTime.Equal(updated.StartTime) {
		t.Errorf("Event time not updated")
	}
}

func TestUpdateEvent_TimingConflict(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()

	// Create two events
	event1 := &storage.Event{
		Title:     "Event 1",
		Username:  "testuser",
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}
	eventID1, _ := s.CreateEvent(ctx, event1)

	event2 := &storage.Event{
		Title:     "Event 2",
		Username:  "testuser",
		StartTime: now.Add(2 * time.Hour),
		EndTime:   now.Add(3 * time.Hour),
	}
	_, _ = s.CreateEvent(ctx, event2)

	// Try to update event1 to overlap with event2
	conflict := &storage.Event{
		Title:     "Conflict",
		Username:  "testuser",
		StartTime: now.Add(2*time.Hour + 30*time.Minute),
		EndTime:   now.Add(4 * time.Hour),
	}
	err := s.UpdateEvent(ctx, "testuser", eventID1, conflict)
	if !errors.Is(err, storage.ErrDateBusy) {
		t.Errorf("Expected ErrDateBusy, got %v", err)
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := &storage.Event{
		Title:     "Test",
		Username:  "testuser",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	}

	err := s.UpdateEvent(ctx, "testuser", uuid.New(), event)
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestDeleteEvent_Success(t *testing.T) {
	s := New()
	ctx := context.Background()

	event := &storage.Event{
		Title:     "Test",
		Username:  "testuser",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
	}
	eventID, _ := s.CreateEvent(ctx, event)

	err := s.DeleteEvent(ctx, "testuser", eventID)
	if err != nil {
		t.Fatalf("DeleteEvent failed: %v", err)
	}

	// Verify deletion
	_, err = s.GetEvent(ctx, "testuser", eventID)
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Event should be deleted")
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	s := New()
	ctx := context.Background()

	err := s.DeleteEvent(ctx, "testuser", uuid.New())
	if !errors.Is(err, storage.ErrEventNotFound) {
		t.Errorf("Expected ErrEventNotFound, got %v", err)
	}
}

func TestListEvents_WithinWindow(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	startWindow := now.Add(-1 * time.Hour)
	endWindow := now.Add(3 * time.Hour)

	// Create event within window
	event1 := &storage.Event{
		Title:     "Event 1",
		Username:  "testuser",
		StartTime: now,
		EndTime:   now.Add(1 * time.Hour),
	}
	s.CreateEvent(ctx, event1)

	// Create event outside window
	event2 := &storage.Event{
		Title:     "Event 2",
		Username:  "testuser",
		StartTime: now.Add(4 * time.Hour),
		EndTime:   now.Add(5 * time.Hour),
	}
	s.CreateEvent(ctx, event2)

	events, err := s.ListEvents(ctx, "testuser", startWindow, endWindow)
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Event 1" {
		t.Errorf("Wrong event returned")
	}
}

func TestListEvents_PartialOverlap(t *testing.T) {
	s := New()
	ctx := context.Background()

	now := time.Now()
	startWindow := now.Add(1 * time.Hour)
	endWindow := now.Add(3 * time.Hour)

	// Event starts before window but ends inside
	event := &storage.Event{
		Title:     "Partial",
		Username:  "testuser",
		StartTime: now,
		EndTime:   now.Add(2 * time.Hour),
	}
	s.CreateEvent(ctx, event)

	events, _ := s.ListEvents(ctx, "testuser", startWindow, endWindow)
	if len(events) != 1 {
		t.Errorf("Partial overlap should be included")
	}
}

func TestListEvents_EmptyForNonexistentUser(t *testing.T) {
	s := New()
	ctx := context.Background()

	events, err := s.ListEvents(ctx, "nonexistent", time.Now(), time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("ListEvents should not error for nonexistent user: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected empty list for nonexistent user")
	}
}

func TestConcurrentCreateAndDelete(t *testing.T) {
	s := New()
	ctx := context.Background()
	numIterations := 20

	var wg sync.WaitGroup
	wg.Add(numIterations)
	errChan := make(chan error, numIterations)

	baseTime := time.Now()

	for i := 0; i < numIterations; i++ {
		// Creator and deleter goroutine - each creates event in its own time slot to avoid conflicts
		go func(id int) {
			defer wg.Done()
			// Each event has unique time slot (20 minutes apart to avoid overlaps)
			event := &storage.Event{
				Title:     "Event",
				Username:  "testuser",
				StartTime: baseTime.Add(time.Duration(id*20) * time.Minute),
				EndTime:   baseTime.Add(time.Duration(id*20+10) * time.Minute),
			}
			eventID, err := s.CreateEvent(ctx, event)
			if err != nil {
				errChan <- err
				return
			}

			// Small delay to ensure deletion happens
			time.Sleep(10 * time.Millisecond)

			// Try to delete
			if err := s.DeleteEvent(ctx, "testuser", eventID); err != nil {
				errChan <- err
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent create/delete error: %v", err)
	}
}

func TestConcurrentUpdateSameEvent(t *testing.T) {
	s := New()
	ctx := context.Background()

	baseTime := time.Now()
	event := &storage.Event{
		Title:     "Event",
		Username:  "testuser",
		StartTime: baseTime,
		EndTime:   baseTime.Add(1 * time.Hour),
	}
	eventID, _ := s.CreateEvent(ctx, event)

	var wg sync.WaitGroup
	numUpdaters := 20
	wg.Add(numUpdaters)
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numUpdaters; i++ {
		go func(id int) {
			defer wg.Done()
			updated := &storage.Event{
				Title:     "Updated",
				Username:  "testuser",
				StartTime: baseTime.Add(time.Duration((id+100)*10) * time.Minute),
				EndTime:   baseTime.Add(time.Duration((id+100)*10+5) * time.Minute),
			}
			err := s.UpdateEvent(ctx, "testuser", eventID, updated)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// All updates should succeed since they don't conflict
	if successCount != numUpdaters {
		t.Errorf("Expected all updates to succeed, got %d/%d", successCount, numUpdaters)
	}
}
