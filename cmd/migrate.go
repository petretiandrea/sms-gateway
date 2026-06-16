package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type migrateOptions struct {
	databaseURL string
	sourceURL   string
	steps       int
	force       int
}

func newMigrateCommand(log *zap.SugaredLogger) *cobra.Command {
	options := migrateOptions{
		sourceURL: "file://migrations",
	}

	command := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
	}

	command.PersistentFlags().StringVar(&options.databaseURL, "database", "", "Postgres DSN. Defaults to ENV_POSTGRES__DSN or POSTGRES_DSN.")
	command.PersistentFlags().StringVar(&options.sourceURL, "source", options.sourceURL, "Migration source URL.")

	command.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migration, err := newMigration(options)
			if err != nil {
				return err
			}
			defer closeMigration(migration)

			if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			log.Info("database migrations applied")
			return nil
		},
	})

	down := &cobra.Command{
		Use:   "down",
		Short: "Roll back all migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			migration, err := newMigration(options)
			if err != nil {
				return err
			}
			defer closeMigration(migration)

			if err := migration.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			log.Info("database migrations rolled back")
			return nil
		},
	}
	command.AddCommand(down)

	steps := &cobra.Command{
		Use:   "steps",
		Short: "Apply N migration steps. Use negative values to roll back.",
		RunE: func(cmd *cobra.Command, args []string) error {
			migration, err := newMigration(options)
			if err != nil {
				return err
			}
			defer closeMigration(migration)

			if err := migration.Steps(options.steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			log.Infow("database migration steps applied", "steps", options.steps)
			return nil
		},
	}
	steps.Flags().IntVar(&options.steps, "steps", 0, "Number of migration steps to apply.")
	command.AddCommand(steps)

	version := &cobra.Command{
		Use:   "version",
		Short: "Print current migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			migration, err := newMigration(options)
			if err != nil {
				return err
			}
			defer closeMigration(migration)

			version, dirty, err := migration.Version()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "version=%d dirty=%t\n", version, dirty)
			return nil
		},
	}
	command.AddCommand(version)

	force := &cobra.Command{
		Use:   "force",
		Short: "Force migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			migration, err := newMigration(options)
			if err != nil {
				return err
			}
			defer closeMigration(migration)

			if err := migration.Force(options.force); err != nil {
				return err
			}
			log.Infow("database migration version forced", "version", options.force)
			return nil
		},
	}
	force.Flags().IntVar(&options.force, "version", -1, "Version to force.")
	command.AddCommand(force)

	return command
}

func newMigration(options migrateOptions) (*migrate.Migrate, error) {
	databaseURL := options.databaseURL
	if databaseURL == "" {
		databaseURL = os.Getenv("ENV_POSTGRES__DSN")
	}
	if databaseURL == "" {
		databaseURL = os.Getenv("POSTGRES_DSN")
	}
	if databaseURL == "" {
		return nil, errors.New("postgres dsn is required")
	}
	if options.sourceURL == "" {
		return nil, errors.New("migration source is required")
	}

	migration, err := migrate.New(options.sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create migration runner: %w", err)
	}

	return migration, nil
}

func closeMigration(migration *migrate.Migrate) {
	sourceErr, databaseErr := migration.Close()
	if sourceErr != nil || databaseErr != nil {
		zap.L().Warn("failed to close migration resources", zap.Error(errors.Join(sourceErr, databaseErr)))
	}
}
