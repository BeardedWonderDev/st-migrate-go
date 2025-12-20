package migration

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"testing"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/stretchr/testify/require"
)

type stubSource struct {
	firstFn    func() (uint, error)
	nextFn     func(uint) (uint, error)
	readUpFn   func(uint) (io.ReadCloser, string, error)
	readDownFn func(uint) (io.ReadCloser, string, error)
}

func (s stubSource) Open(url string) (source.Driver, error) { return s, nil }
func (s stubSource) Close() error                          { return nil }
func (s stubSource) First() (uint, error)                  { return s.firstFn() }
func (s stubSource) Prev(uint) (uint, error)               { return 0, fs.ErrNotExist }
func (s stubSource) Next(v uint) (uint, error)             { return s.nextFn(v) }
func (s stubSource) ReadUp(v uint) (io.ReadCloser, string, error) {
	return s.readUpFn(v)
}
func (s stubSource) ReadDown(v uint) (io.ReadCloser, string, error) {
	return s.readDownFn(v)
}

type errReadCloser struct {
	err error
}

func (e errReadCloser) Read(_ []byte) (int, error) { return 0, e.err }
func (e errReadCloser) Close() error               { return nil }

func TestLoadAllEmptySource(t *testing.T) {
	src := stubSource{
		firstFn: func() (uint, error) { return 0, fs.ErrNotExist },
	}
	ms, err := LoadAll(src, nil)
	require.NoError(t, err)
	require.Empty(t, ms)
}

func TestLoadAllReadUpError(t *testing.T) {
	src := stubSource{
		firstFn: func() (uint, error) { return 1, nil },
		readUpFn: func(uint) (io.ReadCloser, string, error) {
			return nil, "", errors.New("read up")
		},
		readDownFn: func(uint) (io.ReadCloser, string, error) {
			return io.NopCloser(bytes.NewReader([]byte("down"))), "down", nil
		},
		nextFn: func(uint) (uint, error) { return 0, fs.ErrNotExist },
	}
	_, err := LoadAll(src, nil)
	require.Error(t, err)
}

func TestLoadAllReadUpContentError(t *testing.T) {
	src := stubSource{
		firstFn: func() (uint, error) { return 1, nil },
		readUpFn: func(uint) (io.ReadCloser, string, error) {
			return errReadCloser{err: errors.New("read up content")}, "up", nil
		},
		readDownFn: func(uint) (io.ReadCloser, string, error) {
			return io.NopCloser(bytes.NewReader([]byte("down"))), "down", nil
		},
		nextFn: func(uint) (uint, error) { return 0, fs.ErrNotExist },
	}
	_, err := LoadAll(src, nil)
	require.Error(t, err)
}

func TestLoadAllReadDownError(t *testing.T) {
	src := stubSource{
		firstFn: func() (uint, error) { return 1, nil },
		readUpFn: func(uint) (io.ReadCloser, string, error) {
			return io.NopCloser(bytes.NewReader([]byte("up"))), "up", nil
		},
		readDownFn: func(uint) (io.ReadCloser, string, error) {
			return nil, "", errors.New("read down")
		},
		nextFn: func(uint) (uint, error) { return 0, fs.ErrNotExist },
	}
	_, err := LoadAll(src, nil)
	require.Error(t, err)
}

func TestLoadAllNextError(t *testing.T) {
	src := stubSource{
		firstFn: func() (uint, error) { return 1, nil },
		readUpFn: func(uint) (io.ReadCloser, string, error) {
			return io.NopCloser(bytes.NewReader([]byte("up"))), "up", nil
		},
		readDownFn: func(uint) (io.ReadCloser, string, error) {
			return io.NopCloser(bytes.NewReader([]byte("down"))), "down", nil
		},
		nextFn: func(uint) (uint, error) { return 0, errors.New("next") },
	}
	_, err := LoadAll(src, nil)
	require.Error(t, err)
}
