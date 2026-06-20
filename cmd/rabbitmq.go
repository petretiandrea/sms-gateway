package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	rabbitmqtopology "sms-gateway/internal/infra/messaging/rabbitmq"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	defaultRabbitMQExchange = rabbitmqtopology.ExchangeSMS
)

type rabbitMQTopologyOptions struct {
	dsn           string
	exchange      string
	retries       int
	retryInterval time.Duration
}

func newRabbitMQCommand(log *zap.SugaredLogger) *cobra.Command {
	command := &cobra.Command{
		Use:   "rabbitmq",
		Short: "Manage RabbitMQ resources",
	}

	command.AddCommand(newRabbitMQTopologyCommand(log))

	return command
}

func newRabbitMQTopologyCommand(log *zap.SugaredLogger) *cobra.Command {
	options := rabbitMQTopologyOptions{
		exchange:      defaultRabbitMQExchange,
		retries:       30,
		retryInterval: 2 * time.Second,
	}

	command := &cobra.Command{
		Use:   "topology",
		Short: "Manage RabbitMQ topology",
	}

	apply := &cobra.Command{
		Use:   "apply",
		Short: "Declare RabbitMQ exchanges, queues, and bindings",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRabbitMQTopologyApply(cmd.Context(), options, log)
		},
	}
	apply.Flags().StringVar(&options.dsn, "dsn", "", "RabbitMQ DSN. Defaults to ENV_RABBITMQ__DSN, RABBITMQ_DSN, or RABBITMQ_URL.")
	apply.Flags().StringVar(&options.exchange, "exchange", options.exchange, "RabbitMQ topic exchange name.")
	apply.Flags().IntVar(&options.retries, "retries", options.retries, "Number of attempts before failing.")
	apply.Flags().DurationVar(&options.retryInterval, "retry-interval", options.retryInterval, "Delay between retry attempts.")

	command.AddCommand(apply)

	return command
}

func runRabbitMQTopologyApply(ctx context.Context, options rabbitMQTopologyOptions, log *zap.SugaredLogger) error {
	if options.retries < 1 {
		options.retries = 1
	}

	var lastErr error
	for attempt := 1; attempt <= options.retries; attempt++ {
		lastErr = applyRabbitMQTopology(options)
		if lastErr == nil {
			log.Infow("rabbitmq topology applied", "exchange", options.exchange)
			return nil
		}

		if attempt == options.retries {
			break
		}

		log.Warnw(
			"rabbitmq topology apply failed, retrying",
			"attempt", attempt,
			"retries", options.retries,
			"error", lastErr,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(options.retryInterval):
		}
	}

	return fmt.Errorf("apply rabbitmq topology: %w", lastErr)
}

func applyRabbitMQTopology(options rabbitMQTopologyOptions) error {
	dsn := options.dsn
	if dsn == "" {
		dsn = os.Getenv("ENV_RABBITMQ__DSN")
	}
	if dsn == "" {
		dsn = os.Getenv("RABBITMQ_DSN")
	}
	if dsn == "" {
		dsn = os.Getenv("RABBITMQ_URL")
	}
	if dsn == "" {
		return errors.New("rabbitmq dsn is required")
	}
	if options.exchange == "" {
		return errors.New("rabbitmq exchange is required")
	}

	connection, err := amqp.Dial(dsn)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer channel.Close()

	if err := channel.ExchangeDeclare(
		options.exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange %q: %w", options.exchange, err)
	}

	queues := []struct {
		name      string
		arguments amqp.Table
	}{
		{
			name: rabbitmqtopology.QueueSMSSendInternal,
			arguments: amqp.Table{
				"x-dead-letter-exchange":    options.exchange,
				"x-dead-letter-routing-key": rabbitmqtopology.RoutingKeySMSSendDeadLetter,
			},
		},
		{name: rabbitmqtopology.QueueSMSSendDeadLetter},
		{name: rabbitmqtopology.QueueSMSDeliveryConfirmed},
	}

	for _, queue := range queues {
		if _, err := channel.QueueDeclare(
			queue.name,
			true,
			false,
			false,
			false,
			queue.arguments,
		); err != nil {
			return fmt.Errorf("declare queue %q: %w", queue.name, err)
		}
	}

	bindings := []struct {
		queue      string
		routingKey string
	}{
		{queue: rabbitmqtopology.QueueSMSSendInternal, routingKey: rabbitmqtopology.RoutingKeySMSSendInternalRequested},
		{queue: rabbitmqtopology.QueueSMSSendDeadLetter, routingKey: rabbitmqtopology.RoutingKeySMSSendDeadLetter},
		{queue: rabbitmqtopology.QueueSMSDeliveryConfirmed, routingKey: rabbitmqtopology.RoutingKeySMSDeliveryConfirmed},
	}

	for _, binding := range bindings {
		if err := channel.QueueBind(
			binding.queue,
			binding.routingKey,
			options.exchange,
			false,
			nil,
		); err != nil {
			return fmt.Errorf("bind queue %q with routing key %q: %w", binding.queue, binding.routingKey, err)
		}
	}

	return nil
}
