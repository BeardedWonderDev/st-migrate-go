package file

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
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
		slog.Error("create state directory", slog.String("path", path), slog.Any("err", err))
		return nil, err
	}
	return &Store{path: path}, nil
}

func (s *Store) Version(_ context.Context) (int, bool, error) {
	st, err := s.read()
	if err != nil {
		slog.Error("read state file", slog.String("path", s.path), slog.Any("err", err))
		return 0, false, err
	}
	return st.Version, st.Dirty, nil
}

func (s *Store) SetVersion(_ context.Context, version int, dirty bool) error {
	if err := s.write(state{Version: version, Dirty: dirty}); err != nil {
		slog.Error("write state file", slog.String("path", s.path), slog.Int("version", version), slog.Bool("dirty", dirty), slog.Any("err", err))
		return err
	}
	slog.Info("state updated", slog.String("path", s.path), slog.Int("version", version), slog.Bool("dirty", dirty))
	return nil
}

func (s *Store) Lock(_ context.Context) error {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	if s.locked {
		slog.Warn("state lock already held", slog.String("path", s.path))
		return ErrLocked
	}
	s.locked = true
	slog.Debug("state lock acquired", slog.String("path", s.path))
	return nil
}

func (s *Store) Unlock(_ context.Context) error {
	s.lockMu.Lock()
	defer s.lockMu.Unlock()
	if !s.locked {
		slog.Warn("state unlock requested when not locked", slog.String("path", s.path))
		return ErrNotLocked
	}
	s.locked = false
	slog.Debug("state lock released", slog.String("path", s.path))
	return nil
}

func (s *Store) Close() error { return nil }

func (s *Store) read() (state, error) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Debug("state file missing; defaulting", slog.String("path", s.path))
			return state{Version: 0, Dirty: false}, nil
		}
		slog.Error("read state file", slog.String("path", s.path), slog.Any("err", err))
		return state{}, err
	}
	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		slog.Error("parse state file", slog.String("path", s.path), slog.Any("err", err))
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
		slog.Error("marshal state", slog.Any("err", err))
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		slog.Error("write temp state file", slog.String("tmp", tmp), slog.Any("err", err))
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		slog.Error("rename temp state file", slog.String("tmp", tmp), slog.String("path", s.path), slog.Any("err", err))
		return err
	}
	return nil
}
