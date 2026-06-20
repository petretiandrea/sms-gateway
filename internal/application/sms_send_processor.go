package application

import (
	"context"
	"fmt"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/messages"
)

type SMSPushSender interface {
	Send(ctx context.Context, message domain.Sms, token string) error
}

type SMSSendProcessor struct {
	messages domain.Repository
	phones   domain.PhoneRepository
	push     SMSPushSender
}

func NewSMSSendProcessor(
	messages domain.Repository,
	phones domain.PhoneRepository,
	push SMSPushSender,
) *SMSSendProcessor {
	return &SMSSendProcessor{
		messages: messages,
		phones:   phones,
		push:     push,
	}
}

func (processor *SMSSendProcessor) Process(ctx context.Context, body []byte) error {
	message, err := messages.UnmarshalSMSSendRequested(body)
	if err != nil {
		return nonRetryableError{err: fmt.Errorf("invalid sms send message: %w", err)}
	}
	if message.Type != messages.MessageTypeSMSSendRequested {
		return nonRetryableError{err: fmt.Errorf("invalid sms send message: unexpected type %q", message.Type)}
	}
	if message.Data.MessageID == "" || message.Data.PhoneID == "" || message.Data.AccountID == "" {
		return nonRetryableError{err: fmt.Errorf("invalid sms send message: messageId, phoneId and accountId are required")}
	}

	sms := processor.messages.FindById(ctx, domain.SmsId(message.Data.MessageID))
	if sms == nil {
		return nonRetryableError{err: fmt.Errorf("sms send reference not found: sms %q", message.Data.MessageID)}
	}
	if string(sms.UserId) != message.Data.AccountID {
		return nonRetryableError{err: fmt.Errorf("invalid sms send message: sms %q does not belong to account %q", sms.Id, message.Data.AccountID)}
	}

	phone := processor.phones.FindById(ctx, domain.PhoneId(message.Data.PhoneID))
	if phone == nil {
		return nonRetryableError{err: fmt.Errorf("sms send reference not found: phone %q", message.Data.PhoneID)}
	}
	if phone.UserId != sms.UserId {
		return nonRetryableError{err: fmt.Errorf("invalid sms send message: phone %q does not belong to account %q", phone.Id, message.Data.AccountID)}
	}
	if phone.Token == "" {
		return nonRetryableError{err: fmt.Errorf("sms send reference not found: phone %q has no fcm token", phone.Id)}
	}

	if err := processor.push.Send(ctx, *sms, string(phone.Token)); err != nil {
		return fmt.Errorf("send firebase push: %w", err)
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
