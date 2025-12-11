package stmigrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

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

	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	store, err := resolveStore(cfg, logger)
	if err != nil {
		return nil, err
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

// NewWithWrappedDatabase builds a migrate driver from the provided *sql.DB and driver name, then constructs a Runner.
// The driver is built with a dedicated connection so closing the runner will not close the caller's DB pool.
func NewWithWrappedDatabase(cfg Config, driverName string, db *sql.DB, tableName string) (*Runner, error) {
	cfg.DB = db
	cfg.DBDriver = driverName
	cfg.MigrationsTable = tableName
	return New(cfg)
}

// NewWithWrappedDriver wraps an existing migrate database.Driver and constructs a Runner.
func NewWithWrappedDriver(cfg Config, driver database.Driver) (*Runner, error) {
	if driver == nil {
		return nil, fmt.Errorf("driver is nil")
	}
	cfg.Store = state.NewMigrateAdapter(driver)
	return New(cfg)
}

func resolveStore(cfg Config, logger *slog.Logger) (state.Store, error) {
	if cfg.Store != nil {
		return cfg.Store, nil
	}
	if cfg.DBDriver != "" {
		if cfg.DB == nil {
			return nil, fmt.Errorf("db driver set but DB is nil")
		}
		table := cfg.MigrationsTable
		if table == "" {
			table = defaultMigrationsTable
		}
		store, err := buildStoreFromDB(strings.ToLower(cfg.DBDriver), cfg.DB, table, cfg.SkipCloseDB, logger)
		if err != nil {
			return nil, err
		}
		return store, nil
	}
	return memory.New(), nil
}

func buildStoreFromDB(driverName string, db *sql.DB, table string, skipClose bool, logger *slog.Logger) (state.Store, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	ctx := context.Background()

	switch driverName {
	case "postgres", "postgresql":
		conn, err := db.Conn(ctx)
		if err != nil {
			logger.Error("open postgres conn", slog.Any("err", err))
			return nil, fmt.Errorf("open postgres conn: %w", err)
		}
		drv, err := postgres.WithConnection(ctx, conn, &postgres.Config{MigrationsTable: table})
		if err != nil {
			logger.Error("create postgres driver", slog.Any("err", err))
			return nil, fmt.Errorf("create postgres driver: %w", err)
		}
		return state.NewMigrateAdapter(drv), nil
	case "mysql":
		conn, err := db.Conn(ctx)
		if err != nil {
			logger.Error("open mysql conn", slog.Any("err", err))
			return nil, fmt.Errorf("open mysql conn: %w", err)
		}
		drv, err := mysql.WithConnection(ctx, conn, &mysql.Config{MigrationsTable: table})
		if err != nil {
			logger.Error("create mysql driver", slog.Any("err", err))
			return nil, fmt.Errorf("create mysql driver: %w", err)
		}
		return state.NewMigrateAdapter(drv), nil
	case "sqlite3":
		drv, err := sqlite3.WithInstance(db, &sqlite3.Config{MigrationsTable: table})
		if err != nil {
			logger.Error("create sqlite driver", slog.Any("err", err))
			return nil, fmt.Errorf("create sqlite driver: %w", err)
		}
		store := state.NewMigrateAdapter(drv)
		if skipClose {
			return state.WrapNoClose(store), nil
		}
		return store, nil
	default:
		return nil, fmt.Errorf("unsupported driver %q", driverName)
	}
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
