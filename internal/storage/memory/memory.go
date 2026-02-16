package memory

import (
	"sync"

	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/storage"
)

type Storage struct {
	mUsers  map[string]storage.User
	muUsers sync.RWMutex

	mEvents  map[uuid.UUID]storage.Event
	muEvents sync.RWMutex
}

func New() *Storage {
	var s Storage
	s.mUsers = make(map[string]storage.User)
	s.mEvents = make(map[uuid.UUID]storage.Event)

	return &s
}
