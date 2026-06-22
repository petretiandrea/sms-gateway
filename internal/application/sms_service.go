package application

import (
	"context"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/messages"
	"time"

	"github.com/google/uuid"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	"github.com/pkg/errors"
)

type SmsService struct {
	phone     PhoneService
	repo      domain.Repository
	uow       domain.UnitOfWork
	publisher outbox.Publisher
}

type CreateMessageCommand struct {
	Content        string
	To             string
	From           string
	Account        domain.UserAccount
	IdempotencyKey string
	WebhookUrl     *string
	Metadata       *map[string]string
}

func NewSmsService(
	repo domain.Repository,
	phoneService PhoneService,
	uow domain.UnitOfWork,
	publisher outbox.Publisher,
) SmsService {
	return SmsService{repo: repo, phone: phoneService, uow: uow, publisher: publisher}
}

func (service *SmsService) SendSMS(ctx context.Context, params CreateMessageCommand) (*domain.Sms, error) {
	var message *domain.Sms
	err := service.uow.Tx(ctx, func(ctx context.Context) error {
		if message = service.repo.FindExisting(ctx, params.IdempotencyKey); message != nil {
			return nil
		} else {
			// retrieve phoneAccount associated
			var metadata map[string]string
			if params.Metadata == nil {
				metadata = make(map[string]string)
			} else {
				metadata = *params.Metadata
			}
			message = domain.CreateNewSMS(
				params.Account.Id,
				domain.PhoneNumber{Number: params.From},
				domain.PhoneNumber{Number: params.To},
				params.Content,
				params.IdempotencyKey,
				metadata,
				domain.WebhookConfiguration{Url: params.WebhookUrl},
			)
			phoneAccount, err := service.phone.GetPhoneByNumber(ctx, message.From)
			if err != nil {
				return err
			}
			if phoneAccount == nil {
				return nil
			}
			_, err = service.repo.Save(ctx, message)
			if err != nil {
				return err
			}
			return service.publisher.Publish(ctx, messages.NewSMSSendRequested(
				uuid.NewString(),
				time.Now(),
				string(message.Id),
			))
		}
	})
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (service *SmsService) GetSMS(ctx context.Context, id domain.SmsId) *domain.Sms {
	return service.repo.FindById(ctx, id)
}

func (service *SmsService) RegisterAttempt(
	ctx context.Context,
	id domain.SmsId,
	accountID domain.AccountID,
	attempt domain.Attempt,
) (*domain.Sms, error) {
	var message *domain.Sms
	err := service.uow.Tx(ctx, func(ctx context.Context) error {
		message = service.repo.FindById(ctx, id)
		if message == nil {
			return nil
		}
		if message.UserId != accountID {
			return errors.Wrapf(
				domain.ErrorNotMessageOwner,
				"Not owner of sms [%s]",
				message.Id,
			)
		}

		message.RegisterAttempt(attempt)
		save, err := service.repo.Save(ctx, message)
		if err != nil {
			return err
		}

		message = save
		return service.publisher.Publish(ctx, messages.NewSMSAttemptRegistered(
			uuid.NewString(),
			time.Now(),
			string(message.Id),
		))
	})
	if err != nil {
		return nil, err
	}
	return message, nil
}
