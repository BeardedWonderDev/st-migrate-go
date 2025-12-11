package stmigrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/BeardedWonderDev/st-migrate-go/internal/migration"
	"github.com/BeardedWonderDev/st-migrate-go/internal/schema"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state/memory"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
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
		logger.Debug("no source url provided; using default", slog.String("source", sourceURL))
	}

	src, err := source.Open(sourceURL)
	if err != nil {
		logger.Error("open source", slog.String("source", sourceURL), slog.Any("err", err))
		return nil, fmt.Errorf("open source %s: %w", sourceURL, err)
	}
	migrations, err := migration.LoadAll(src, logger)
	src.Close()
	if err != nil {
		logger.Error("load migrations", slog.String("source", sourceURL), slog.Any("err", err))
		return nil, fmt.Errorf("load migrations: %w", err)
	}

	logger.Info("constructed runner",
		slog.String("source", sourceURL),
		slog.Bool("dry_run", cfg.DryRun),
		slog.String("executor", fmt.Sprintf("%T", exec)),
		slog.String("store", fmt.Sprintf("%T", store)),
	)

	r := migration.NewRunner(store, exec, reg, logger, cfg.DryRun, migrations)
	return &Runner{inner: r}, nil
}

const defaultMigrationsTable = "st_schema_migrations"

// WrapMigrateDatabase constructs a golang-migrate SQL driver with a default migrations table name
// of "st_schema_migrations" (overridable via tableName) and wraps it as a state.Store.
// Supported drivers: postgres, mysql, sqlite3.
func WrapMigrateDatabase(driverName string, db *sql.DB, tableName string) (state.Store, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	if tableName == "" {
		tableName = defaultMigrationsTable
	}

	var drv database.Driver
	var err error
	switch driverName {
	case "postgres", "postgresql":
		drv, err = postgres.WithInstance(db, &postgres.Config{MigrationsTable: tableName})
	case "mysql":
		drv, err = mysql.WithInstance(db, &mysql.Config{MigrationsTable: tableName})
	case "sqlite3":
		drv, err = sqlite3.WithInstance(db, &sqlite3.Config{MigrationsTable: tableName})
	default:
		return nil, fmt.Errorf("unsupported driver %q", driverName)
	}
	if err != nil {
		return nil, fmt.Errorf("create migrate driver: %w", err)
	}
	return state.NewMigrateAdapter(drv), nil
}

// WrapMigrateDriver wraps an already-constructed golang-migrate database driver as a state.Store.
// Prefer WrapMigrateDatabase to control migrations table naming automatically - Use only for go-lang-migrate datbases not already implemented.
func WrapMigrateDriver(driver database.Driver) state.Store {
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
