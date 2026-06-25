package application

import (
	"context"
	"sms-gateway/internal/domain"

	outboxamqp "github.com/petretiandrea/outbox-go/pkg/outbox/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

type DLQConsumer struct {
	repo domain.DLQMessageRepository
}

func NewDLQConsumer(repo domain.DLQMessageRepository) *DLQConsumer {
	return &DLQConsumer{repo: repo}
}

func (consumer *DLQConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	message := outboxamqp.MessageFromDelivery(delivery)

	return consumer.repo.Save(ctx, domain.DLQMessage{
		MessageID:   message.ID,
		Channel:     string(message.Channel),
		Exchange:    delivery.Exchange,
		RoutingKey:  delivery.RoutingKey,
		ContentType: delivery.ContentType,
		Payload:     []byte(message.Payload),
		Metadata:    message.Metadata,
		OccurredAt:  message.OccurredAt,
	})
}
