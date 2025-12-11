package sdk

import (
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
}
