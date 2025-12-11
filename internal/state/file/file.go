package file

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var (
	// ErrLocked signals the store is already locked.
	ErrLocked = errors.New("state store locked")
	// ErrNotLocked signals unlock without prior lock.
	ErrNotLocked = errors.New("state store not locked")
)

type Store struct {
	path    string
	lockMu  sync.Mutex
	locked  bool
	stateMu sync.Mutex
}

type state struct {
	Version int  `json:"version"`
	Dirty   bool `json:"dirty"`
}

// New creates a file-backed store. The path will be created if missing.
func New(path string) (*Store, error) {
	if path == "" {
		path = ".st-migrate/state.json"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	return &Store{path: path}, nil
}

func (s *Store) Version(_ context.Context) (int, bool, error) {
	st, err := s.read()
	if err != nil {
		return 0, false, err
	}
	return st.Version, st.Dirty, nil
}

func (s *Store) SetVersion(_ context.Context, version int, dirty bool) error {
	return s.write(state{Version: version, Dirty: dirty})
}

func (s *Store) Lock(_ context.Context) error {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	if s.locked {
		return ErrLocked
	}
	s.locked = true
	return nil
}

func (s *Store) Unlock(_ context.Context) error {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	if !s.locked {
		return ErrNotLocked
	}
	s.locked = false
	return nil
}

func (s *Store) Close() error { return nil }

func (s *Store) read() (state, error) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return state{Version: 0, Dirty: false}, nil
		}
		return state{}, err
	}
	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return state{}, err
	}
	return st, nil
}

func (s *Store) write(st state) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	tmp := s.path + ".tmp"
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
