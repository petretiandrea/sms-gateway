package application

import (
	"context"
	"testing"
	"time"

	"sms-gateway/internal/domain"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/messages"

	"github.com/google/uuid"
)

type fakePhoneRepository struct {
	phone *domain.Phone
}

func (repo fakePhoneRepository) Save(ctx context.Context, phone domain.Phone) (*domain.Phone, error) {
	return &phone, nil
}

func (repo fakePhoneRepository) FindById(ctx context.Context, id domain.PhoneId) *domain.Phone {
	if repo.phone != nil && repo.phone.Id == id {
		return repo.phone
	}
	return nil
}

func (repo fakePhoneRepository) FindByPhoneNumber(ctx context.Context, number domain.PhoneNumber) *domain.Phone {
	if repo.phone != nil && repo.phone.Phone == number {
		return repo.phone
	}
	return nil
}

func (repo fakePhoneRepository) Delete(ctx context.Context, id domain.PhoneId) bool {
	return false
}

type fakeSMSAttemptRegistrar struct {
	called    bool
	smsID     domain.SmsId
	accountID domain.AccountID
	attempt   domain.Attempt
}

func (registrar *fakeSMSAttemptRegistrar) RegisterAttempt(
	ctx context.Context,
	id domain.SmsId,
	accountID domain.AccountID,
	attempt domain.Attempt,
) (*domain.Sms, error) {
	registrar.called = true
	registrar.smsID = id
	registrar.accountID = accountID
	registrar.attempt = attempt
	return &domain.Sms{Id: id, UserId: accountID, LastAttempt: attempt}, nil
}

func TestSMSSendProcessorDryRunRegistersSuccessfulAttemptWithoutFCMToken(t *testing.T) {
	accountID := domain.AccountID(uuid.NewString())
	phone := &domain.Phone{
		Id:     domain.PhoneId(uuid.NewString()),
		Phone:  domain.PhoneNumber{Number: "+390000000000"},
		UserId: accountID,
		Token:  "",
	}
	sms := &domain.Sms{
		Id:           domain.SmsId(uuid.NewString()),
		From:         phone.Phone,
		To:           "+391111111111",
		UserId:       accountID,
		CreatedAt:    time.Now(),
		LastUpdateAt: time.Now(),
	}
	push := infra.NewFirebasePushNotification(nil)
	push.EnableDryRun()
	registrar := &fakeSMSAttemptRegistrar{}
	processor := NewSMSSendProcessor(fakeSmsRepository{sms: sms}, fakePhoneRepository{phone: phone}, &push, registrar)

	message, err := messages.NewSMSSendRequested("evt-1", time.Now(), string(sms.Id))
	if err != nil {
		t.Fatalf("NewSMSSendRequested() error = %v", err)
	}
	if err := processor.Handle(context.Background(), message); err != nil {
		t.Fatalf("processor.Handle() error = %v", err)
	}

	if !registrar.called {
		t.Fatal("expected attempt registrar to be called")
	}
	if registrar.smsID != sms.Id {
		t.Fatalf("registrar.smsID = %q, want %q", registrar.smsID, sms.Id)
	}
	if registrar.accountID != accountID {
		t.Fatalf("registrar.accountID = %q, want %q", registrar.accountID, accountID)
	}
	success, ok := registrar.attempt.(domain.SuccessAttempt)
	if !ok {
		t.Fatalf("registrar.attempt = %T, want domain.SuccessAttempt", registrar.attempt)
	}
	if success.AttemptCount != 1 {
		t.Fatalf("success.AttemptCount = %d, want 1", success.AttemptCount)
	}
	if success.PhoneId != phone.Id {
		t.Fatalf("success.PhoneId = %q, want %q", success.PhoneId, phone.Id)
	}
}
