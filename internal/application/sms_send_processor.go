package application

import (
	"context"
	"fmt"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/messages"

	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

type SMSAttemptRegistrar interface {
	RegisterAttempt(ctx context.Context, id domain.SmsId, accountID domain.AccountID, attempt domain.Attempt) (*domain.Sms, error)
}

type SMSSendProcessor struct {
	messages domain.Repository
	phones   domain.PhoneRepository
	push     *infra.FirebasePushNotification
	attempts SMSAttemptRegistrar
}

func NewSMSSendProcessor(
	messages domain.Repository,
	phones domain.PhoneRepository,
	push *infra.FirebasePushNotification,
	attempts SMSAttemptRegistrar,
) *SMSSendProcessor {
	return &SMSSendProcessor{
		messages: messages,
		phones:   phones,
		push:     push,
		attempts: attempts,
	}
}

func (processor *SMSSendProcessor) Handle(ctx context.Context, message outbox.Message) error {
	requested, err := messages.UnmarshalSMSSendRequested(message.Payload)
	if err != nil {
		return nonRetryableError{err: fmt.Errorf("invalid sms send message: %w", err)}
	}

	if requested.MessageID == "" {
		return nonRetryableError{err: fmt.Errorf("invalid sms send message: messageId is required")}
	}

	sms := processor.messages.FindById(ctx, domain.SmsId(requested.MessageID))
	if sms == nil {
		return nonRetryableError{err: fmt.Errorf("sms %q not found", requested.MessageID)}
	}

	phone := processor.phones.FindByPhoneNumber(ctx, sms.From)
	if phone == nil {
		return nonRetryableError{err: fmt.Errorf("phone for sms %q not found", sms.Id)}
	}

	if phone.Token == "" && !processor.push.IsDryRun() {
		return nonRetryableError{err: fmt.Errorf("phone %q has no fcm token", phone.Id)}
	}

	if err := processor.push.Send(ctx, *sms, string(phone.Token)); err != nil {
		return err
	}

	if processor.push.IsDryRun() {
		_, err := processor.attempts.RegisterAttempt(ctx, sms.Id, sms.UserId, domain.SuccessAttempt{
			AttemptCount: 1,
			PhoneId:      phone.Id,
		})
		return err
	}

	return nil
}

type nonRetryableError struct {
	err error
}

func (err nonRetryableError) Error() string {
	return err.err.Error()
}

func (err nonRetryableError) Unwrap() error {
	return err.err
}

func (err nonRetryableError) Retryable() bool {
	return false
}
