package migration

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/BeardedWonderDev/st-migrate-go/internal/schema"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state"
)

// Runner coordinates applying migrations using a state store and executor.
type Runner struct {
	store      state.Store
	exec       executor.Executor
	registry   *schema.Registry
	logger     *slog.Logger
	dryRun     bool
	migrations []Migration
}

// NewRunner constructs a Runner with parsed migrations.
func NewRunner(store state.Store, exec executor.Executor, registry *schema.Registry, logger *slog.Logger, dryRun bool, migrations []Migration) *Runner {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return &Runner{
		store:      store,
		exec:       exec,
		registry:   registry,
		logger:     logger,
		dryRun:     dryRun,
		migrations: sortMigrations(migrations),
	}
}

// Up applies pending migrations up to the optional target version.
// If target is nil, all pending migrations are applied.
func (r *Runner) Up(ctx context.Context, target *uint) error {
	r.logger.Info("up start", slog.Any("target", target), slog.Bool("dry_run", r.dryRun), slog.Int("available", len(r.migrations)))
	if err := r.store.Lock(ctx); err != nil {
		r.logger.Error("lock state store", slog.Any("err", err))
		return fmt.Errorf("lock state store: %w", err)
	}
	defer func() {
		if err := r.store.Unlock(ctx); err != nil {
			r.logger.Error("unlock state store", slog.Any("err", err))
		}
	}()

	current, dirty, err := r.store.Version(ctx)
	if err != nil {
		r.logger.Error("read version", slog.Any("err", err))
		return err
	}
	if current < 0 {
		r.logger.Debug("normalizing negative current version to zero", slog.Int("current", current))
		current = 0
	}
	if dirty {
		r.logger.Warn("state is dirty; refusing to run migrations")
		return fmt.Errorf("state is dirty; resolve before running migrations")
	}

	applied := 0
	for _, m := range r.migrations {
		if target != nil && m.Version > *target {
			break
		}
		if int(m.Version) <= current {
			r.logger.Debug("skip already applied", slog.Uint64("version", uint64(m.Version)))
			continue
		}
		if err := r.apply(ctx, m.Version, m.Up); err != nil {
			_ = r.store.SetVersion(ctx, int(m.Version), true)
			r.logger.Error("apply up migration", slog.Uint64("version", uint64(m.Version)), slog.Any("err", err))
			return err
		}
		if r.dryRun {
			continue
		}
		if err := r.store.SetVersion(ctx, int(m.Version), false); err != nil {
			r.logger.Error("persist version", slog.Uint64("version", uint64(m.Version)), slog.Any("err", err))
			return err
		}
		r.logger.Info("applied migration", slog.Uint64("version", uint64(m.Version)), slog.String("direction", "up"))
		applied++
	}
	r.logger.Info("up complete", slog.Int("applied", applied), slog.Any("target", target))
	return nil
}

// Down rolls back a number of migrations (default 1 if steps<=0).
func (r *Runner) Down(ctx context.Context, steps int) error {
	if steps <= 0 {
		steps = 1
	}
	r.logger.Info("down start", slog.Int("steps", steps), slog.Bool("dry_run", r.dryRun))
	if err := r.store.Lock(ctx); err != nil {
		r.logger.Error("lock state store", slog.Any("err", err))
		return fmt.Errorf("lock state store: %w", err)
	}
	defer func() {
		if err := r.store.Unlock(ctx); err != nil {
			r.logger.Error("unlock state store", slog.Any("err", err))
		}
	}()

	current, dirty, err := r.store.Version(ctx)
	if err != nil {
		r.logger.Error("read version", slog.Any("err", err))
		return err
	}
	if dirty {
		r.logger.Warn("state is dirty; refusing to run migrations")
		return fmt.Errorf("state is dirty; resolve before running migrations")
	}
	if current <= 0 {
		r.logger.Info("no migrations to roll back")
		return nil
	}

	idx := indexByVersion(r.migrations)
	for i := 0; i < steps && current > 0; i++ {
		m, ok := idx[uint(current)]
		if !ok {
			r.logger.Warn("migration version not found for down", slog.Int("current", current))
			return fmt.Errorf("migration version %d not found for down", current)
		}
		if err := r.apply(ctx, m.Version, m.Down); err != nil {
			_ = r.store.SetVersion(ctx, int(m.Version), true)
			r.logger.Error("apply down migration", slog.Uint64("version", uint64(m.Version)), slog.Any("err", err))
			return err
		}
		if r.dryRun {
			continue
		}
		prev := previousVersion(r.migrations, m.Version)
		if err := r.store.SetVersion(ctx, int(prev), false); err != nil {
			r.logger.Error("persist version", slog.Uint64("version", uint64(prev)), slog.Any("err", err))
			return err
		}
		r.logger.Info("rolled back migration", slog.Uint64("version", uint64(m.Version)))
		current = int(prev)
	}
	r.logger.Info("down complete", slog.Int("steps_requested", steps), slog.Int("current_version", current))
	return nil
}

// Status reports the current applied version and the pending migration versions.
func (r *Runner) Status(ctx context.Context) (int, []uint, error) {
	current, _, err := r.store.Version(ctx)
	if err != nil {
		r.logger.Error("read version", slog.Any("err", err))
		return 0, nil, err
	}
	pending := make([]uint, 0)
	for _, m := range r.migrations {
		if int(m.Version) > current {
			pending = append(pending, m.Version)
		}
	}
	return current, pending, nil
}

// Close releases resources on the store, if any.
func (r *Runner) Close() error {
	return r.store.Close()
}

// Migrate moves to the target version, applying up or down as needed.
func (r *Runner) Migrate(ctx context.Context, target uint) error {
	r.logger.Info("migrate start", slog.Uint64("target", uint64(target)), slog.Bool("dry_run", r.dryRun))
	if err := r.store.Lock(ctx); err != nil {
		r.logger.Error("lock state store", slog.Any("err", err))
		return fmt.Errorf("lock state store: %w", err)
	}
	defer func() {
		if err := r.store.Unlock(ctx); err != nil {
			r.logger.Error("unlock state store", slog.Any("err", err))
		}
	}()

	current, dirty, err := r.store.Version(ctx)
	if err != nil {
		r.logger.Error("read version", slog.Any("err", err))
		return err
	}
	if current < 0 {
		r.logger.Debug("normalizing negative current version to zero", slog.Int("current", current))
		current = 0
	}
	if dirty {
		r.logger.Warn("state is dirty; refusing to migrate")
		return fmt.Errorf("state is dirty; resolve before running migrations")
	}
	if uint(current) == target {
		r.logger.Info("no-op migrate; already at target", slog.Int("current", current))
		return nil
	}

	maxVersion := r.migrations[len(r.migrations)-1].Version
	if target > maxVersion {
		r.logger.Error("target version not found", slog.Uint64("target", uint64(target)), slog.Uint64("max_available", uint64(maxVersion)))
		return fmt.Errorf("target version %d not found; max available %d", target, maxVersion)
	}

	if target > uint(current) {
		return r.upTo(ctx, target)
	}
	return r.downTo(ctx, target, uint(current))
}

func (r *Runner) upTo(ctx context.Context, target uint) error {
	r.logger.Debug("migrate up path", slog.Uint64("target", uint64(target)))
	for _, m := range r.migrations {
		if m.Version > target {
			break
		}
		if err := r.apply(ctx, m.Version, m.Up); err != nil {
			_ = r.store.SetVersion(ctx, int(m.Version), true)
			r.logger.Error("apply up migration", slog.Uint64("version", uint64(m.Version)), slog.Any("err", err))
			return err
		}
		if r.dryRun {
			continue
		}
		if err := r.store.SetVersion(ctx, int(m.Version), false); err != nil {
			r.logger.Error("persist version", slog.Uint64("version", uint64(m.Version)), slog.Any("err", err))
			return err
		}
		r.logger.Info("applied migration", slog.Uint64("version", uint64(m.Version)), slog.String("direction", "up"))
	}
	r.logger.Info("migrate up complete", slog.Uint64("target", uint64(target)))
	return nil
}

func (r *Runner) downTo(ctx context.Context, target uint, current uint) error {
	r.logger.Debug("migrate down path", slog.Uint64("target", uint64(target)), slog.Uint64("current", uint64(current)))
	// Build a set of versions we need to roll back.
	needed := make(map[uint]Migration)
	for _, m := range r.migrations {
		if m.Version > target && m.Version <= current {
			needed[m.Version] = m
		}
	}
	// Roll back from current downward.
	v := current
	for v > target {
		m, ok := needed[v]
		if !ok {
			r.logger.Warn("missing migration for rollback", slog.Uint64("version", uint64(v)))
			return fmt.Errorf("missing migration version %d for rollback", v)
		}
		if err := r.apply(ctx, m.Version, m.Down); err != nil {
			_ = r.store.SetVersion(ctx, int(m.Version), true)
			r.logger.Error("apply down migration", slog.Uint64("version", uint64(m.Version)), slog.Any("err", err))
			return err
		}
		if !r.dryRun {
			prev := previousVersion(r.migrations, m.Version)
			if err := r.store.SetVersion(ctx, int(prev), false); err != nil {
				r.logger.Error("persist version", slog.Uint64("version", uint64(prev)), slog.Any("err", err))
				return err
			}
			r.logger.Info("rolled back migration", slog.Uint64("version", uint64(m.Version)))
			v = prev
		} else {
			v = previousVersion(r.migrations, m.Version)
		}
	}
	return nil
}

func (r *Runner) apply(ctx context.Context, version uint, data []byte) error {
	spec, err := r.registry.Parse(data)
	if err != nil {
		r.logger.Error("parse migration", slog.Uint64("version", uint64(version)), slog.Any("err", err))
		return fmt.Errorf("parse migration %d: %w", version, err)
	}
	if r.dryRun {
		r.logger.Info("dry run: would apply migration", slog.Uint64("version", uint64(version)), slog.Int("actions", len(spec.Actions)))
		return nil
	}
	r.logger.Debug("apply migration", slog.Uint64("version", uint64(version)), slog.Int("actions", len(spec.Actions)))
	for _, action := range spec.Actions {
		r.logger.Debug("apply action", slog.String("role", action.Role), slog.String("ensure", action.Ensure), slog.Int("add_count", len(action.Add)), slog.Int("remove_count", len(action.Remove)))
		switch action.Ensure {
		case "present":
			if err := r.exec.EnsureRole(ctx, action.Role); err != nil {
				r.logger.Error("ensure role", slog.String("role", action.Role), slog.Any("err", err))
				return fmt.Errorf("ensure role %s: %w", action.Role, err)
			}
			if len(action.Add) > 0 {
				if err := r.exec.AddPermissions(ctx, action.Role, action.Add); err != nil {
					r.logger.Error("add permissions", slog.String("role", action.Role), slog.Any("err", err))
					return fmt.Errorf("add permissions to %s: %w", action.Role, err)
				}
			}
			if len(action.Remove) > 0 {
				if err := r.exec.RemovePermissions(ctx, action.Role, action.Remove); err != nil {
					r.logger.Error("remove permissions", slog.String("role", action.Role), slog.Any("err", err))
					return fmt.Errorf("remove permissions from %s: %w", action.Role, err)
				}
			}
		case "absent":
			if err := r.exec.DeleteRole(ctx, action.Role); err != nil {
				r.logger.Error("delete role", slog.String("role", action.Role), slog.Any("err", err))
				return fmt.Errorf("delete role %s: %w", action.Role, err)
			}
		default:
			r.logger.Error("unknown ensure value", slog.String("ensure", action.Ensure), slog.String("role", action.Role))
			return fmt.Errorf("unknown ensure value %q for role %s", action.Ensure, action.Role)
		}
	}
	return nil
}

func sortMigrations(ms []Migration) []Migration {
	out := make([]Migration, len(ms))
	copy(out, ms)
	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out
}

func indexByVersion(ms []Migration) map[uint]Migration {
	result := make(map[uint]Migration, len(ms))
	for _, m := range ms {
		result[m.Version] = m
	}
	return result
}

func previousVersion(ms []Migration, version uint) uint {
	prev := uint(0)
	for _, m := range ms {
		if m.Version < version && m.Version > prev {
			prev = m.Version
		}
	}
	return prev
}
