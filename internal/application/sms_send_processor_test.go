package application

import (
	"context"
	"errors"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/messages"
	"testing"
	"time"
)

type fakeMessageRepository struct {
	message *domain.Sms
}

func (repo fakeMessageRepository) Save(context.Context, domain.Sms) (*domain.Sms, error) {
	return nil, nil
}

func (repo fakeMessageRepository) FindById(context.Context, domain.SmsId) *domain.Sms {
	return repo.message
}

func (repo fakeMessageRepository) FindExisting(context.Context, string) *domain.Sms {
	return nil
}

func (repo fakeMessageRepository) Find(context.Context, domain.QueryParams) ([]domain.Sms, error) {
	return nil, nil
}

type fakePhoneRepository struct {
	phone *domain.Phone
}

func (repo fakePhoneRepository) Save(context.Context, domain.Phone) (*domain.Phone, error) {
	return nil, nil
}

func (repo fakePhoneRepository) FindById(context.Context, domain.PhoneId) *domain.Phone {
	return repo.phone
}

func (repo fakePhoneRepository) FindByPhoneNumber(context.Context, domain.PhoneNumber) *domain.Phone {
	return nil
}

func (repo fakePhoneRepository) Delete(context.Context, domain.PhoneId) bool {
	return false
}

type fakePushSender struct {
	called bool
	err    error
}

func (sender *fakePushSender) Send(_ context.Context, _ domain.Sms, _ string) error {
	sender.called = true
	return sender.err
}

func TestSMSSendProcessorSendsPush(t *testing.T) {
	body := mustSMSSendRequested(t, "message-1", "phone-1", "account-1")
	push := &fakePushSender{}
	processor := NewSMSSendProcessor(
		fakeMessageRepository{message: &domain.Sms{Id: "message-1", UserId: "account-1"}},
		fakePhoneRepository{phone: &domain.Phone{Id: "phone-1", UserId: "account-1", Token: "token-1"}},
		push,
	)

	if err := processor.Process(context.Background(), body); err != nil {
		t.Fatalf("process sms send requested: %v", err)
	}
	if !push.called {
		t.Fatal("expected push sender to be called")
	}
}

func TestSMSSendProcessorRejectsUnexpectedType(t *testing.T) {
	body := []byte(`{"id":"event-1","type":"unknown","occurredAt":"2026-06-20T09:00:00Z","aggregateId":"message-1","data":{"messageId":"message-1","phoneId":"phone-1","accountId":"account-1"}}`)
	processor := NewSMSSendProcessor(fakeMessageRepository{}, fakePhoneRepository{}, &fakePushSender{})

	err := processor.Process(context.Background(), body)
	var retryable interface{ Retryable() bool }
	if !errors.As(err, &retryable) || retryable.Retryable() {
		t.Fatalf("expected non-retryable error, got %v", err)
	}
}

func TestSMSSendProcessorReturnsSenderError(t *testing.T) {
	body := mustSMSSendRequested(t, "message-1", "phone-1", "account-1")
	expectedErr := errors.New("firebase unavailable")
	push := &fakePushSender{err: expectedErr}
	processor := NewSMSSendProcessor(
		fakeMessageRepository{message: &domain.Sms{Id: "message-1", UserId: "account-1"}},
		fakePhoneRepository{phone: &domain.Phone{Id: "phone-1", UserId: "account-1", Token: "token-1"}},
		push,
	)

	err := processor.Process(context.Background(), body)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected sender error, got %v", err)
	}
}

func mustSMSSendRequested(t *testing.T, messageID string, phoneID string, accountID string) []byte {
	t.Helper()
	message := messages.NewSMSSendRequested("event-1", time.Date(2026, 6, 20, 9, 0, 0, 0, time.UTC), messageID, phoneID, accountID)
	body, err := messages.MarshalEnvelope(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	return body
}
