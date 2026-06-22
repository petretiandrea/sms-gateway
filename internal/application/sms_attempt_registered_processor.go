package application

import (
	"context"
	"fmt"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/messages"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

type SMSDeliveryNotifier interface {
	NotifyDelivery(ctx context.Context, sms domain.Sms) error
}

type SMSAttemptRegisteredProcessor struct {
	messages domain.Repository
	notifier SMSDeliveryNotifier
}

func NewSMSAttemptRegisteredProcessor(
	messages domain.Repository,
	notifier SMSDeliveryNotifier,
) *SMSAttemptRegisteredProcessor {
	return &SMSAttemptRegisteredProcessor{
		messages: messages,
		notifier: notifier,
	}
}

func (processor *SMSAttemptRegisteredProcessor) Handle(ctx context.Context, message outbox.Message) error {
	requested, err := messages.UnmarshalSMSAttemptRegistered(message.Payload)
	if err != nil {
		return nonRetryableError{err: fmt.Errorf("invalid sms attempt registered message: %w", err)}
	}

	if requested.MessageID == "" {
		return nonRetryableError{err: fmt.Errorf("invalid sms attempt registered message: messageId is required")}
	}

	sms := processor.messages.FindById(ctx, domain.SmsId(requested.MessageID))
	if sms == nil {
		return nonRetryableError{err: fmt.Errorf("sms %q not found", requested.MessageID)}
	}

	return processor.notifier.NotifyDelivery(ctx, *sms)
}
