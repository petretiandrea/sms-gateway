package application

import (
	"context"
	"fmt"
	"sms-gateway/internal/infra/messaging/rabbitmq"
	"sms-gateway/internal/messages"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
	outboxamqp "github.com/petretiandrea/outbox-go/pkg/outbox/amqp"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SMSSendRequestedHandler interface {
	Handle(ctx context.Context, message outbox.Message) error
}

type SMSAttemptRegisteredHandler interface {
	Handle(ctx context.Context, message outbox.Message) error
}

type SMSOutboxConsumer struct {
	smsSendProcessor              SMSSendRequestedHandler
	smsAttemptRegisteredProcessor SMSAttemptRegisteredHandler
}

func NewSMSOutboxConsumer(
	smsSendProcessor SMSSendRequestedHandler,
	smsAttemptRegisteredProcessor SMSAttemptRegisteredHandler,
) *SMSOutboxConsumer {
	return &SMSOutboxConsumer{
		smsSendProcessor:              smsSendProcessor,
		smsAttemptRegisteredProcessor: smsAttemptRegisteredProcessor,
	}
}

func (consumer *SMSOutboxConsumer) Process(ctx context.Context, delivery amqp.Delivery) error {
	message := outboxamqp.MessageFromDelivery(delivery)
	switch message.Channel {
	case messages.ChannelSMSSendInternal:
		return consumer.smsSendProcessor.Handle(ctx, message)
	case messages.ChannelSMSAttemptInternal:
		return consumer.smsAttemptRegisteredProcessor.Handle(ctx, message)
	default:
		return nonRetryableError{err: fmt.Errorf("unsupported outbox channel %q", message.Channel)}
	}
}

var _ rabbitmq.Handler = (*SMSOutboxConsumer)(nil)
