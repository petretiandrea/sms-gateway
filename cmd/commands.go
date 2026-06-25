package main

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newRootCommand(version string, log *zap.SugaredLogger) *cobra.Command {
	root := &cobra.Command{
		Use:     "sms-gateway",
		Short:   "SMS Gateway service",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd.Context(), version, log)
		},
	}

	root.AddCommand(newStartCommand(version, log))
	root.AddCommand(newMigrateCommand(log))
	root.AddCommand(newRabbitMQCommand(log))
	root.AddCommand(newMongoDBCommand(log))

	return root
}

func newStartCommand(version string, log *zap.SugaredLogger) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the HTTP service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd.Context(), version, log)
		},
	}
}

func runStart(ctx context.Context, version string, log *zap.SugaredLogger) error {
	container := NewContainer(ctx, version, log)
	defer container.Close()

	log.Infof("Running SMS Gateway, version: %s\n", version)

	return container.StartHTTPServer()
}
