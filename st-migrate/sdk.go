package stmigrate

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/BeardedWonderDev/st-migrate-go/internal/migration"
	"github.com/BeardedWonderDev/st-migrate-go/internal/schema"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state/memory"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source"
	// register default file source driver
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Runner is the public API surface for applying migrations.
type Runner struct {
	inner *migration.Runner
}

var defaultExecutorFactory = func() executor.Executor {
	return executor.NewSuperTokensExecutor()
}

// SetDefaultExecutorFactory overrides the default executor factory (intended for tests).
func SetDefaultExecutorFactory(f func() executor.Executor) {
	if f != nil {
		defaultExecutorFactory = f
	}
}

// New constructs a Runner using the provided configuration.
// If Store is nil, an in-memory store is used. If Executor is nil, SuperTokens is used.
// If Registry is nil, the default schema registry is used.
func New(cfg Config) (*Runner, error) {
	reg := cfg.Registry
	if reg == nil {
		reg = schema.DefaultRegistry()
	}

	exec := cfg.Executor
	if exec == nil {
		exec = defaultExecutorFactory()
	}

	store := cfg.Store
	if store == nil {
		store = memory.New()
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	sourceURL := cfg.SourceURL
	if sourceURL == "" {
		sourceURL = "file://backend/migrations/auth"
	}

	src, err := source.Open(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("open source %s: %w", sourceURL, err)
	}
	migrations, err := migration.LoadAll(src)
	src.Close()
	if err != nil {
		return nil, fmt.Errorf("load migrations: %w", err)
	}

	r := migration.NewRunner(store, exec, reg, logger, cfg.DryRun, migrations)
	return &Runner{inner: r}, nil
}

// WrapMigrateDatabase allows consumers to supply a golang-migrate database driver as the state store.
func WrapMigrateDatabase(driver database.Driver) state.Store {
	if driver == nil {
		return nil
	}
	return state.NewMigrateAdapter(driver)
}

func (r *Runner) Up(ctx context.Context, target *uint) error {
	return r.inner.Up(ctx, target)
}

func (r *Runner) Down(ctx context.Context, steps int) error {
	return r.inner.Down(ctx, steps)
}

func (r *Runner) Status(ctx context.Context) (int, []uint, error) {
	return r.inner.Status(ctx)
}

func (r *Runner) Close() error {
	return r.inner.Close()
}

// Migrate moves to the target version, applying up or down as needed.
func (r *Runner) Migrate(ctx context.Context, target uint) error {
	return r.inner.Migrate(ctx, target)
}
