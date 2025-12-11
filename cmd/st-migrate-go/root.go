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

	return rootCmd
}

func buildRunner(opts *cliOpts) (*stmigrate.Runner, error) {
	level := slog.LevelInfo
	if opts.verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(opts.output, &slog.HandlerOptions{Level: level}))

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
		cfg.Store = stmigrate.WrapMigrateDatabase(drv)
	} else {
		fs, err := filestore.New(opts.stateFile)
		if err != nil {
			return nil, fmt.Errorf("init state file: %w", err)
		}
		cfg.Store = fs
	}

	return stmigrate.New(cfg)
}

func upCmd(opts *cliOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "up [target]",
		Short: "Apply pending migrations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var target *uint
			if len(args) == 1 {
				n, err := strconv.ParseUint(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid target version: %w", err)
				}
				val := uint(n)
				target = &val
			}
			runner, err := buildRunner(opts)
			if err != nil {
				return err
			}
			defer runner.Close()
			return runner.Up(context.Background(), target)
		},
	}
}

func downCmd(opts *cliOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "down [steps]",
		Short: "Roll back applied migrations",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			steps := 1
			if len(args) == 1 {
				n, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid steps: %w", err)
				}
				steps = n
			}
			runner, err := buildRunner(opts)
			if err != nil {
				return err
			}
			defer runner.Close()
			return runner.Down(context.Background(), steps)
		},
	}
}

func statusCmd(opts *cliOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current and pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner, err := buildRunner(opts)
			if err != nil {
				return err
			}
			defer runner.Close()
			current, pending, err := runner.Status(context.Background())
			if err != nil {
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

func createCmd(opts *cliOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create paired up/down migration files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := sourceURLToPath(opts.sourceURL)
			opts := create.Options{
				Dir:           dir,
				Name:          args[0],
				Width:         opts.width,
				SchemaVersion: opts.schemaVer,
			}
			up, down, err := create.Scaffold(opts)
			if err != nil {
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
