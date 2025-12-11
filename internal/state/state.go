package state

import "context"

// Store tracks migration state (current version and dirty flag).
// This mirrors the golang-migrate database.Driver version/locking surface
// so existing drivers can be wrapped with minimal shims.
type Store interface {
	Version(ctx context.Context) (version int, dirty bool, err error)
	SetVersion(ctx context.Context, version int, dirty bool) error
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
	Close() error
}
