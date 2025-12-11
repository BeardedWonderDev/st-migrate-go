package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/BeardedWonderDev/st-migrate-go/internal/create"
	"github.com/BeardedWonderDev/st-migrate-go/sdk"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/spf13/cobra"
)

var (
	sourceURL   string
	databaseURL string
	dryRun      bool
	verbose     bool
	width       int
	schemaVer   int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "st-migrate-go",
		Short: "Role/permission migration runner",
	}

	rootCmd.PersistentFlags().StringVar(&sourceURL, "source", "file://backend/migrations/auth", "migration source URL (golang-migrate style)")
	rootCmd.PersistentFlags().StringVar(&databaseURL, "database", "", "state database URL (golang-migrate driver)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "print actions without executing")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose logging")

	rootCmd.AddCommand(upCmd())
	rootCmd.AddCommand(downCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(createCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func buildRunner() (*sdk.Runner, error) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	cfg := sdk.Config{
		SourceURL: sourceURL,
		DryRun:    dryRun,
		Logger:    logger,
	}

	if databaseURL != "" {
		drv, err := database.Open(databaseURL)
		if err != nil {
			return nil, fmt.Errorf("open database driver: %w", err)
		}
		cfg.Store = sdk.WrapMigrateDatabase(drv)
	}

	return sdk.New(cfg)
}

func upCmd() *cobra.Command {
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
			runner, err := buildRunner()
			if err != nil {
				return err
			}
			defer runner.Close()
			return runner.Up(context.Background(), target)
		},
	}
}

func downCmd() *cobra.Command {
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
			runner, err := buildRunner()
			if err != nil {
				return err
			}
			defer runner.Close()
			return runner.Down(context.Background(), steps)
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current and pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			runner, err := buildRunner()
			if err != nil {
				return err
			}
			defer runner.Close()
			current, pending, err := runner.Status(context.Background())
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "current version: %d\n", current)
			if len(pending) == 0 {
				fmt.Fprintln(os.Stdout, "pending: none")
			} else {
				fmt.Fprintf(os.Stdout, "pending: %v\n", pending)
			}
			return nil
		},
	}
}

func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create paired up/down migration files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := create.Options{
				Dir:           sourceURLToPath(sourceURL),
				Name:          args[0],
				Width:         width,
				SchemaVersion: schemaVer,
			}
			up, down, err := create.Scaffold(opts)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "created %s\ncreated %s\n", up, down)
			return nil
		},
	}
	cmd.Flags().IntVar(&width, "digits", 4, "zero-pad width for version numbers")
	cmd.Flags().IntVar(&schemaVer, "schema-version", 1, "schema version to use in generated files")
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
