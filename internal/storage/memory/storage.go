package memorystorage

import (
	"sync"

	"github.com/mrvin/hw-otus-go/hw12-15calendar/internal/storage"
)

type Storage struct {
	mUsers  map[string]storage.User
	muUsers sync.RWMutex

	mEvents    map[int64]storage.Event
	maxIDEvent int64
	muEvents   sync.RWMutex
}

func New() *Storage {
	var s Storage
	s.mUsers = make(map[string]storage.User)
	s.mEvents = make(map[int64]storage.Event)

	return &s
}
