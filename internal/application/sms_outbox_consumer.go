package application

import (
	"context"
	"fmt"
	"sms-gateway/internal/infra/messaging/rabbitmq"
	"sms-gateway/internal/messages"
	"strings"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
	outboxamqp "github.com/petretiandrea/outbox-go/pkg/outbox/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SMSSendRequestedHandler interface {
	Handle(ctx context.Context, message outbox.Message) error
}

type SMSOutboxConsumer struct {
	smsSendProcessor SMSSendRequestedHandler
}

func NewSMSOutboxConsumer(smsSendProcessor SMSSendRequestedHandler) *SMSOutboxConsumer {
	return &SMSOutboxConsumer{
		smsSendProcessor: smsSendProcessor,
	}
}

func (consumer *SMSOutboxConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	message := outboxamqp.MessageFromDelivery(delivery)
	switch strings.TrimSpace(message.Metadata["type"]) {
	case messages.MessageTypeSMSSendRequested:
		return consumer.smsSendProcessor.Handle(ctx, message)
	default:
		return nonRetryableError{err: fmt.Errorf("unsupported outbox message type %q", message.Metadata["type"])}
	}
}

var _ rabbitmq.Handler = (*SMSOutboxConsumer)(nil)
