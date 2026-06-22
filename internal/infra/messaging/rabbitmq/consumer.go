package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Handler interface {
	Process(ctx context.Context, delivery amqp.Delivery) error
}

type retryableError interface {
	Retryable() bool
}

type Consumer struct {
	dsn         string
	queue       string
	handler     Handler
	log         *zap.Logger
	connection  *amqp.Connection
	channel     *amqp.Channel
	stopConsume context.CancelFunc
}

func NewConsumer(dsn string, queue string, handler Handler, log *zap.Logger) *Consumer {
	if log == nil {
		log = zap.NewNop()
	}

	return &Consumer{
		dsn:     dsn,
		queue:   queue,
		handler: handler,
		log:     log.Named("rabbitmq_consumer").With(zap.String("queue", queue)),
	}
}

func (consumer *Consumer) Start(ctx context.Context) {
	for {
		if err := consumer.consume(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			consumer.log.Error("rabbitmq consumer stopped, retrying", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (consumer *Consumer) Stop() {
	if consumer.stopConsume != nil {
		consumer.stopConsume()
	}
	if consumer.channel != nil {
		if err := consumer.channel.Close(); err != nil {
			consumer.log.Warn("failed to close rabbitmq channel", zap.Error(err))
		}
	}
	if consumer.connection != nil {
		if err := consumer.connection.Close(); err != nil {
			consumer.log.Warn("failed to close rabbitmq connection", zap.Error(err))
		}
	}
}

func (consumer *Consumer) consume(ctx context.Context) error {
	connection, err := amqp.Dial(consumer.dsn)
	if err != nil {
		return fmt.Errorf("connect rabbitmq: %w", err)
	}
	consumer.connection = connection

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	consumer.channel = channel

	if err := channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("configure rabbitmq qos: %w", err)
	}

	consumeCtx, cancel := context.WithCancel(ctx)
	consumer.stopConsume = cancel
	deliveries, err := channel.ConsumeWithContext(
		consumeCtx,
		consumer.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume queue %q: %w", consumer.queue, err)
	}

	consumer.log.Info("rabbitmq consumer started")
	for delivery := range deliveries {
		if err := consumer.handler.Process(ctx, delivery); err != nil {
			consumer.log.Error("failed to process rabbitmq message", zap.Error(err))
			if isRetryable(err) {
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Reject(false)
			continue
		}

		if err := delivery.Ack(false); err != nil {
			consumer.log.Warn("failed to ack rabbitmq message", zap.Error(err))
		}
	}

	return nil
}

func isRetryable(err error) bool {
	var retryable retryableError
	if errors.As(err, &retryable) {
		return retryable.Retryable()
	}
	return true
}
