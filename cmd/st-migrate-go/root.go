package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/BeardedWonderDev/st-migrate-go/internal/create"
	filestore "github.com/BeardedWonderDev/st-migrate-go/internal/state/file"
	"github.com/BeardedWonderDev/st-migrate-go/st-migrate"
	"github.com/golang-migrate/migrate/v4/database"
	// common database drivers registered for CLI
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/spf13/cobra"
)

type cliOpts struct {
	sourceURL string
	database  string
	stateFile string
	dryRun    bool
	verbose   bool
	width     int
	schemaVer int
	output    io.Writer
	logger    *slog.Logger
}

func newRootCmd(out io.Writer) *cobra.Command {
	opts := cliOpts{
		sourceURL: "file://backend/migrations/auth",
		stateFile: ".st-migrate/state.json",
		width:     4,
		schemaVer: 1,
		output:    out,
	}

	rootCmd := &cobra.Command{
		Use:   "st-migrate-go",
		Short: "Role/permission migration runner for SuperTokens",
	}

	rootCmd.PersistentFlags().StringVar(&opts.sourceURL, "source", opts.sourceURL, "migration source URL (golang-migrate style)")
	rootCmd.PersistentFlags().StringVar(&opts.database, "database", "", "state database URL (golang-migrate driver)")
	rootCmd.PersistentFlags().StringVar(&opts.stateFile, "state-file", opts.stateFile, "path to file-based state store (used when --database is empty)")
	rootCmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", false, "print actions without executing")
	rootCmd.PersistentFlags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logging")

	rootCmd.AddCommand(upCmd(&opts))
	rootCmd.AddCommand(downCmd(&opts))
	rootCmd.AddCommand(statusCmd(&opts))
	rootCmd.AddCommand(createCmd(&opts))
	rootCmd.AddCommand(migrateCmd(&opts))

	return rootCmd
}

func buildRunner(opts *cliOpts) (*stmigrate.Runner, error) {
	logger := getLogger(opts)

	cfg := stmigrate.Config{
		SourceURL: opts.sourceURL,
		DryRun:    opts.dryRun,
		Logger:    logger,
	}

	if opts.database != "" {
		drv, err := database.Open(opts.database)
		if err != nil {
			return nil, fmt.Errorf("open database driver: %w", err)
		}
		cfg.Store = stmigrate.WrapMigrateDriver(drv)
	} else {
		fs, err := filestore.New(opts.stateFile)
		if err != nil {
			logger.Error("init state file", slog.String("state_file", opts.stateFile), slog.Any("err", err))
			return nil, fmt.Errorf("init state file: %w", err)
		}
		cfg.Store = fs
	}

	return stmigrate.New(cfg)
}

func getLogger(opts *cliOpts) *slog.Logger {
	if opts.logger != nil {
		return opts.logger
	}
	level := slog.LevelInfo
	if opts.verbose {
		level = slog.LevelDebug
	}
	opts.logger = slog.New(slog.NewTextHandler(opts.output, &slog.HandlerOptions{Level: level}))
	return opts.logger
}

func upCmd(opts *cliOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "up [target]",
		Short: "Apply pending migrations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(opts)
			var target *uint
			if len(args) == 1 {
				n, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					logger.Error("invalid target version", slog.String("input", args[0]), slog.Any("err", err))
					return fmt.Errorf("invalid target version: %w", err)
				}
				val := uint(n)
				target = &val
			}
			logger.Info("command: up", slog.String("source", opts.sourceURL), slog.String("database", opts.database), slog.String("state_file", opts.stateFile), slog.Bool("dry_run", opts.dryRun), slog.Any("target", target))
			runner, err := buildRunner(opts)
			if err != nil {
				logger.Error("build runner", slog.Any("err", err))
				return err
			}
			defer runner.Close()
			if err := runner.Up(context.Background(), target); err != nil {
				logger.Error("up failed", slog.Any("err", err))
				return err
			}
			return nil
		},
	}
}

func downCmd(opts *cliOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "down [steps]",
		Short: "Roll back applied migrations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(opts)
			steps := 1
			if len(args) == 1 {
				n, err := strconv.Atoi(args[0])
				if err != nil {
					logger.Error("invalid steps", slog.String("input", args[0]), slog.Any("err", err))
					return fmt.Errorf("invalid steps: %w", err)
				}
				steps = n
			}
			logger.Info("command: down", slog.String("source", opts.sourceURL), slog.String("database", opts.database), slog.String("state_file", opts.stateFile), slog.Bool("dry_run", opts.dryRun), slog.Int("steps", steps))
			runner, err := buildRunner(opts)
			if err != nil {
				logger.Error("build runner", slog.Any("err", err))
				return err
			}
			defer runner.Close()
			if err := runner.Down(context.Background(), steps); err != nil {
				logger.Error("down failed", slog.Any("err", err))
				return err
			}
			return nil
		},
	}
}

func statusCmd(opts *cliOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current and pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(opts)
			logger.Info("command: status", slog.String("source", opts.sourceURL), slog.String("database", opts.database), slog.String("state_file", opts.stateFile), slog.Bool("dry_run", opts.dryRun))
			runner, err := buildRunner(opts)
			if err != nil {
				logger.Error("build runner", slog.Any("err", err))
				return err
			}
			defer runner.Close()
			current, pending, err := runner.Status(context.Background())
			if err != nil {
				logger.Error("status failed", slog.Any("err", err))
				return err
			}
			fmt.Fprintf(opts.output, "current version: %d\n", current)
			if len(pending) == 0 {
				fmt.Fprintln(opts.output, "pending: none")
			} else {
				fmt.Fprintf(opts.output, "pending: %v\n", pending)
			}
			return nil
		},
	}
}

func migrateCmd(opts *cliOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate <version>",
		Short: "Migrate up or down to the target version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(opts)
			n, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				logger.Error("invalid target version", slog.String("input", args[0]), slog.Any("err", err))
				return fmt.Errorf("invalid target version: %w", err)
			}
			target := uint(n)
			logger.Info("command: migrate", slog.String("source", opts.sourceURL), slog.String("database", opts.database), slog.String("state_file", opts.stateFile), slog.Bool("dry_run", opts.dryRun), slog.Uint64("target", uint64(target)))
			runner, err := buildRunner(opts)
			if err != nil {
				logger.Error("build runner", slog.Any("err", err))
				return err
			}
			defer runner.Close()
			if err := runner.Migrate(context.Background(), target); err != nil {
				logger.Error("migrate failed", slog.Any("err", err))
				return err
			}
			return nil
		},
	}
}

func createCmd(opts *cliOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create paired up/down migration files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := getLogger(opts)
			dir := sourceURLToPath(opts.sourceURL)
			opts := create.Options{
				Dir:           dir,
				Name:          args[0],
				Width:         opts.width,
				SchemaVersion: opts.schemaVer,
			}
			logger.Info("command: create", slog.String("dir", dir), slog.String("name", args[0]), slog.Int("width", opts.Width), slog.Int("schema_version", opts.SchemaVersion))
			up, down, err := create.Scaffold(opts)
			if err != nil {
				logger.Error("create scaffold failed", slog.Any("err", err))
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\ncreated %s\n", up, down)
			return nil
		},
	}
	cmd.Flags().IntVar(&opts.width, "digits", opts.width, "zero-pad width for version numbers")
	cmd.Flags().IntVar(&opts.schemaVer, "schema-version", opts.schemaVer, "schema version to use in generated files")
	return cmd
}

// sourceURLToPath converts a file:// URL to a local path for create scaffolding.
func sourceURLToPath(url string) string {
	const prefix = "file://"
	if len(url) >= len(prefix) && url[:len(prefix)] == prefix {
		return url[len(prefix):]
	}
	return url
}
