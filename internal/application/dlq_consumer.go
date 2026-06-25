package application

import (
	"context"
	"sms-gateway/internal/domain"

	outboxamqp "github.com/petretiandrea/outbox-go/pkg/outbox/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DLQConsumer struct {
	repo    domain.DLQMessageRepository
	metrics DLQMetrics
}

func NewDLQConsumer(repo domain.DLQMessageRepository, metrics DLQMetrics) *DLQConsumer {
	if metrics == nil {
		metrics = noopDLQMetrics{}
	}
	return &DLQConsumer{repo: repo, metrics: metrics}
}

func (consumer *DLQConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	message := outboxamqp.MessageFromDelivery(delivery)

	dlqMessage := domain.DLQMessage{
		MessageID:   message.ID,
		Channel:     string(message.Channel),
		Exchange:    delivery.Exchange,
		RoutingKey:  delivery.RoutingKey,
		ContentType: delivery.ContentType,
		Payload:     []byte(message.Payload),
		Metadata:    message.Metadata,
		OccurredAt:  message.OccurredAt,
	}

	if err := consumer.repo.Save(ctx, dlqMessage); err != nil {
		consumer.metrics.StoreFailed(ctx, dlqMessage.Channel, dlqMessage.RoutingKey)
		return err
	}

	consumer.metrics.MessageStored(ctx, dlqMessage.Channel, dlqMessage.RoutingKey)
	return nil
}
