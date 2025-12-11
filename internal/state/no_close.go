package state

import "context"

// NoCloseStore wraps a Store and makes Close a no-op (useful when the underlying
// driver is shared and should be closed by the caller).
type NoCloseStore struct {
	inner Store
}

func (n NoCloseStore) Version(ctx context.Context) (int, bool, error) {
	return n.inner.Version(ctx)
}

func (n NoCloseStore) SetVersion(ctx context.Context, version int, dirty bool) error {
	return n.inner.SetVersion(ctx, version, dirty)
}

func (n NoCloseStore) Lock(ctx context.Context) error {
	return n.inner.Lock(ctx)
}

func (n NoCloseStore) Unlock(ctx context.Context) error {
	return n.inner.Unlock(ctx)
}

func (n NoCloseStore) Close() error { return nil }

func WrapNoClose(inner Store) Store {
	if inner == nil {
		return nil
	}
	return NoCloseStore{inner: inner}
}
