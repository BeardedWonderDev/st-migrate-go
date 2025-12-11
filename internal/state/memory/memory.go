package memory

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

var (
	// ErrLocked indicates the store is already locked.
	ErrLocked = errors.New("state store locked")
	// ErrNotLocked indicates an unlock was attempted without a lock.
	ErrNotLocked = errors.New("state store not locked")
)

// Store is an in-memory implementation of the state store.
type Store struct {
	mu      sync.Mutex
	locked  bool
	version int
	dirty   bool
}

// New creates a new in-memory store with version 0.
func New() *Store {
	return &Store{}
}

func (s *Store) Version(_ context.Context) (int, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version, s.dirty, nil
}

func (s *Store) SetVersion(_ context.Context, version int, dirty bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.version = version
	s.dirty = dirty
	slog.Info("state updated (memory)", slog.Int("version", version), slog.Bool("dirty", dirty))
	return nil
}

func (s *Store) Lock(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.locked {
		slog.Warn("state lock already held (memory)")
		return ErrLocked
	}
	s.locked = true
	slog.Debug("state lock acquired (memory)")
	return nil
}

func (s *Store) Unlock(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.locked {
		slog.Warn("state unlock requested when not locked (memory)")
		return ErrNotLocked
	}
	s.locked = false
	slog.Debug("state lock released (memory)")
	return nil
}

func (s *Store) Close() error { return nil }
