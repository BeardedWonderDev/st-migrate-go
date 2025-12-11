package stmigrate

import (
	"database/sql"
	"log/slog"

	"github.com/BeardedWonderDev/st-migrate-go/internal/executor"
	"github.com/BeardedWonderDev/st-migrate-go/internal/schema"
	"github.com/BeardedWonderDev/st-migrate-go/internal/state"
)

// Config drives construction of a Runner instance.
type Config struct {
	SourceURL string
	Store     state.Store
	Executor  executor.Executor
	Logger    *slog.Logger
	DryRun    bool
	Registry  *schema.Registry
	// Optional DB parameters to let the SDK build a dedicated migrate driver.
	DB              *sql.DB
	DBDriver        string // postgres | mysql | sqlite3
	MigrationsTable string
}
