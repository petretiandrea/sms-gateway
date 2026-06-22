package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"sms-gateway/internal/domain"
	"sms-gateway/internal/messages"

	"github.com/google/uuid"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
)

type fakeSmsRepository struct {
	sms *domain.Sms
}

func (repo fakeSmsRepository) Save(ctx context.Context, message *domain.Sms) (*domain.Sms, error) {
	return message, nil
}

func (repo fakeSmsRepository) FindById(ctx context.Context, id domain.SmsId) *domain.Sms {
	if repo.sms != nil && repo.sms.Id == id {
		return repo.sms
	}
	return nil
}

func (repo fakeSmsRepository) FindExisting(ctx context.Context, idempotencyKey string) *domain.Sms {
	return nil
}

func (repo fakeSmsRepository) Find(ctx context.Context, params domain.QueryParams) ([]domain.Sms, error) {
	return nil, nil
}

type fakeDeliveryNotifier struct {
	called bool
	smsID  domain.SmsId
}

func (notifier *fakeDeliveryNotifier) NotifyDelivery(ctx context.Context, sms domain.Sms) error {
	notifier.called = true
	notifier.smsID = sms.Id
	return nil
}

func TestSMSAttemptRegisteredProcessorNotifiesDelivery(t *testing.T) {
	expectedSMS := &domain.Sms{
		Id:           domain.SmsId(uuid.NewString()),
		UserId:       domain.AccountID(uuid.NewString()),
		CreatedAt:    time.Now(),
		LastUpdateAt: time.Now(),
	}
	repo := fakeSmsRepository{sms: expectedSMS}
	notifier := &fakeDeliveryNotifier{}
	processor := NewSMSAttemptRegisteredProcessor(repo, notifier)

	message, err := messages.NewSMSAttemptRegistered("evt-1", time.Now(), string(expectedSMS.Id))
	if err != nil {
		t.Fatalf("NewSMSAttemptRegistered() error = %v", err)
	}
	if err := processor.Handle(context.Background(), message); err != nil {
		t.Fatalf("processor.Handle() error = %v", err)
	}

	if !notifier.called {
		t.Fatal("expected notifier to be called")
	}
	if notifier.smsID != expectedSMS.Id {
		t.Fatalf("notifier.smsID = %q, want %q", notifier.smsID, expectedSMS.Id)
	}
}

func TestSMSAttemptRegisteredProcessorRejectsMissingSms(t *testing.T) {
	processor := NewSMSAttemptRegisteredProcessor(fakeSmsRepository{}, &fakeDeliveryNotifier{})

	message := outbox.Message{
		Channel: messages.ChannelSMSAttemptInternal,
		Payload: []byte(`{"messageId":"missing"}`),
	}

	err := processor.Handle(context.Background(), message)
	if err == nil {
		t.Fatal("expected error")
	}

	var retryable interface{ Retryable() bool }
	if !errors.As(err, &retryable) || retryable.Retryable() {
		t.Fatalf("expected non-retryable error, got %v", err)
	}
}
